use std::sync::Arc;
use std::time::Duration;

use futures::StreamExt;
use lazy_static::lazy_static;
use log::{debug, error};
use regex::Regex;
use reqwest::StatusCode;
use serde_derive::{Deserialize, Serialize};
use serde_with::skip_serializing_none;

use crate::replacer::StringReplacer;

/// Follows links, and then runs an optional replacement at the end
#[skip_serializing_none]
#[derive(Deserialize, Serialize, Debug)]
pub struct LinkFollowReplacer {
    name: String,

    #[serde(with = "serde_regex")]
    match_regex: Regex,

    /// `destination_regex` requires `destination_replacement` be set.
    #[serde(with = "serde_regex")]
    #[serde(default)]
    destination_regex: Option<Regex>,

    #[serde(skip_serializing_if = "Option::is_none")]
    destination_replacement: Option<String>,

    /// replace keys with values after link has been followed.
    #[serde(skip_serializing_if = "Option::is_none")]
    post_replacement: Option<Vec<Replacements>>,

    /// if the destination matches this regex it won't be included in the final message.    
    ///
    /// takes priority over destination_replacement and post_replacement, so if this applies those are skipped.
    #[serde(with = "serde_regex")]
    #[serde(default)]
    bad_url_match: Option<Regex>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct Replacements {
    replace: String,
    with: String,
}

lazy_static! {
    static ref CLIENT: Arc<reqwest::Client> = Arc::new(
        reqwest::Client::builder()
            .timeout(Duration::from_secs(10))
            .user_agent("curl/4.0")
            .build()
            .unwrap()
    );
}

impl LinkFollowReplacer {
    #[allow(dead_code)]
    pub fn new(
        name: String,
        match_regex: Regex,
        destination_regex: Option<Regex>,
        destination_replacement: Option<String>,
        post_replacement: Option<Vec<Replacements>>,
        bad_url_match: Option<Regex>,
    ) -> Self {
        Self {
            name,
            match_regex,
            destination_regex,
            destination_replacement,
            post_replacement,
            bad_url_match,
        }
    }

    /// Apply `bad_url_match`, then `destination_regex` or `post_replacement`,
    /// to a link that has already been followed to its destination.
    ///
    /// Returns `None` if the destination matched `bad_url_match`.
    fn replace_destination(&self, link: &str) -> Option<String> {
        if let Some(bad_url_regex) = &self.bad_url_match {
            if bad_url_regex.is_match(link) {
                debug!("{} matched bad_url_regex, skipping link", link);
                return None;
            }
        }

        let mut link = link.to_string();

        if let Some(replacement_regex) = &self.destination_regex {
            let replacement = &self.destination_replacement.clone().unwrap();
            link = replacement_regex.replace_all(&link, replacement).to_string();
        } else if let Some(replacements) = &self.post_replacement {
            for replacement in replacements {
                if link.contains(&replacement.replace) {
                    debug!(
                        "replacing {} with {} in {}",
                        &replacement.replace, &replacement.with, &link
                    );
                    link = link.replace(replacement.replace.as_str(), replacement.with.as_str());
                    debug!("replaced: {}", &link);
                }
            }
        }

        Some(link)
    }
}

async fn visit_links(links: Vec<String>) -> Vec<anyhow::Result<String>> {
    futures::stream::iter(links.into_iter().map(|link| async move {
        let response = CLIENT.get(&link).send().await?;
        if response.status() != StatusCode::OK {
            error!(
                "status code not OK for link {}: {}",
                &link,
                response.status().to_string()
            );
        }

        let mut url = response.url().clone();
        url.set_query(None);
        Ok(url.to_string())
    }))
    .buffer_unordered(8)
    .collect::<Vec<anyhow::Result<String>>>()
    .await
}

impl StringReplacer for LinkFollowReplacer {
    fn matches(&self, message: &str) -> bool {
        self.match_regex.is_match(message)
    }

    async fn replace(&mut self, message: &str) -> anyhow::Result<String> {
        if !self.matches(message) {
            return Ok(message.to_string());
        }

        let links: Vec<String> = self
            .match_regex
            .find_iter(message)
            .map(|link| link.as_str().to_string())
            .collect();

        let results = visit_links(links).await;

        let mut visited_links: Vec<String> = Vec::new();

        for result in results {
            match result {
                Ok(gotten_link) => {
                    if let Some(link) = self.replace_destination(&gotten_link) {
                        visited_links.push(link);
                    }
                }
                Err(err) => error!("error visiting link: {}", err),
            }
        }

        Ok(visited_links.join("\n"))
    }

    fn name(&self) -> &String {
        &self.name
    }
}

#[cfg(test)]
mod tests {
    use std::sync::Arc;

    use super::LinkFollowReplacer;
    use crate::replacer::{get_replacer_by_name, Replacer, StringReplacer};

    fn link_follower(name: &str) -> LinkFollowReplacer {
        let config = Arc::new(crate::config::create().unwrap());
        match get_replacer_by_name(&name.to_string(), config).unwrap() {
            Some(Replacer::LinkReplacer(r)) => r,
            _ => panic!("no link-follower replacer named {}", name),
        }
    }

    #[test]
    fn reddit_share_only_matches_app_share_links() {
        let replacer = link_follower("reddit_share");

        assert!(replacer.matches("https://reddit.com/r/pchelp/s/GPxofho2iY/"));
        assert!(replacer.matches("https://www.reddit.com/r/computers/s/BN4uCFXFyC"));
        assert!(!replacer.matches(
            "https://www.reddit.com/r/pchelp/comments/167kev4/i_kinda_messed_up_should_i_redo_it/"
        ));
    }

    #[test]
    fn reddit_share_rewrites_destination_to_vxreddit() {
        let replacer = link_follower("reddit_share");

        assert_eq!(
            replacer.replace_destination(
                "https://www.reddit.com/r/pchelp/comments/167kev4/i_kinda_messed_up_should_i_redo_it/"
            ),
            Some(
                "https://vxreddit.com/r/pchelp/comments/167kev4/i_kinda_messed_up_should_i_redo_it/"
                    .to_string()
            )
        );
        // reddit bounces links it doesn't want to serve to an interstitial,
        // which must never be posted in place of the original link
        assert_eq!(
            replacer.replace_destination("https://www.reddit.com/login/"),
            None
        );
        assert_eq!(
            replacer.replace_destination("https://old.reddit.com/over18/"),
            None
        );
    }
}

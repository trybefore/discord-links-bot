use std::sync::Arc;
use std::time::Duration;
use anyhow::{anyhow, bail};

use futures::{StreamExt};
use lazy_static::lazy_static;
use log::{debug, error};
use regex::Regex;
use reqwest::{StatusCode};
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
}

async fn visit_links(links: Vec<String>) -> Vec<anyhow::Result<String>> {
    futures::stream::iter(
        links.into_iter().map(|link| {
            async move {
                let response = CLIENT.get(&link)
                    .send()
                    .await?;
                if response.status() != StatusCode::OK {
                    error!("status code not OK for link {}: {}", &link, response.status().to_string());
                }

                let mut url = response.url().clone();
                url.set_query(None);
                Ok(url.to_string())
            }
        })
    ).buffer_unordered(8).collect::<Vec<anyhow::Result<String>>>().await
}

impl StringReplacer for LinkFollowReplacer {
    fn matches(&self, message: &String) -> bool {
        self.match_regex.is_match(message)
    }

    async fn replace(&mut self, message: &String) -> anyhow::Result<String> {
        if !self.matches(message) {
            return Ok(message.to_string());
        }

        let links: Vec<String> = self.match_regex.find_iter(message.as_str()).map(|link| link.as_str().to_string()).collect();

        let results = visit_links(links).await;

        let mut visited_links: Vec<String> = Vec::new();

        for result in results {
            if let Ok(gotten_link) = result {
                let mut link = gotten_link.clone();

                if let Some(bad_url_regex) = &self.bad_url_match {
                    if bad_url_regex.is_match(&link) {
                        debug!("{} matched bad_url_regex, skipping link", &link);
                        continue;
                    }
                }

                if let Some(replacement_regex) = &self.destination_regex {
                    let replacement = &self.destination_replacement.clone().unwrap();
                    link = replacement_regex.replace_all(&link, replacement).to_string();
                } else if let Some(replacements) = &self.post_replacement {
                    for replacement in replacements {
                        if link.contains(&replacement.replace) {
                            debug!("replacing {} with {} in {}", &replacement.replace, &replacement.with, &link);
                            link = link.replace(replacement.replace.as_str(), replacement.with.as_str());
                            debug!("replaced: {}", &link);
                        }
                    }
                }


                visited_links.push(link.to_string());
            } else if let Err(err) = result {
                error!("error visiting link: {}", err);
            }
        }


        Ok(
            visited_links.join("\n")
        )
    }

    fn name(&self) -> &String {
        &self.name
    }
}

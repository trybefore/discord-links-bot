use std::sync::Arc;
use std::time::Duration;

use anyhow::bail;
use futures::StreamExt;
use lazy_static::lazy_static;
use log::{debug, error};
use regex::Regex;
use reqwest::header::LOCATION;
use reqwest::StatusCode;
use serde_derive::{Deserialize, Serialize};
use serde_with::skip_serializing_none;
use url::Url;

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
            // redirects are followed by hand in visit_link, so it can stop
            // before landing on a bad url
            .redirect(reqwest::redirect::Policy::none())
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
            link = replacement_regex
                .replace_all(&link, replacement)
                .to_string();
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

/// Same limit reqwest uses when it follows redirects itself.
const MAX_REDIRECTS: usize = 10;

/// Follow `link` one redirect at a time and return where it ends up, without
/// its query string.
///
/// If a redirect would take us to a URL matching `bad_url_match` (a login or
/// age-gate interstitial), stop and return the last URL before it instead, as
/// that one usually already is the permalink we were looking for. The bad URL
/// is only returned if the very first redirect goes there.
async fn visit_link(link: String, bad_url_match: Option<&Regex>) -> anyhow::Result<String> {
    let mut url = Url::parse(&link)?;

    for hop in 0..=MAX_REDIRECTS {
        let response = CLIENT.get(url.clone()).send().await?;

        let next = if response.status().is_redirection() {
            response
                .headers()
                .get(LOCATION)
                .and_then(|location| location.to_str().ok())
                .and_then(|location| url.join(location).ok())
        } else {
            None
        };

        let Some(next) = next else {
            if response.status() != StatusCode::OK {
                error!(
                    "status code not OK for link {}: {}",
                    &link,
                    response.status().to_string()
                );
            }
            break;
        };

        if hop == MAX_REDIRECTS {
            bail!("too many redirects for link {}", &link);
        }

        if hop > 0 && bad_url_match.is_some_and(|bad| bad.is_match(next.as_str())) {
            debug!(
                "{} redirects to bad url {}, keeping the former",
                url.as_str(),
                next.as_str()
            );
            break;
        }

        url = next;
    }

    url.set_query(None);
    Ok(url.to_string())
}

async fn visit_links(
    links: Vec<String>,
    bad_url_match: Option<&Regex>,
) -> Vec<anyhow::Result<String>> {
    futures::stream::iter(
        links
            .into_iter()
            .map(|link| visit_link(link, bad_url_match)),
    )
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

        let results = visit_links(links, self.bad_url_match.as_ref()).await;

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
    use std::io::{BufRead, BufReader, Write};
    use std::net::TcpListener;
    use std::sync::Arc;

    use regex::Regex;

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

    /// Serve a reddit-like redirect chain on localhost and return its base url:
    /// `/share` -> `/permalink?share_id=1` -> `/login/?dest=...`, and
    /// `/straight-to-login` -> `/login/`.
    fn redirect_server() -> String {
        let listener = TcpListener::bind("127.0.0.1:0").unwrap();
        let base = format!("http://{}", listener.local_addr().unwrap());

        std::thread::spawn(move || {
            for stream in listener.incoming() {
                let mut stream = stream.unwrap();
                let mut request_line = String::new();
                BufReader::new(&stream)
                    .read_line(&mut request_line)
                    .unwrap();
                let path = request_line.split_whitespace().nth(1).unwrap_or("/");

                let location = match path.split('?').next().unwrap() {
                    "/share" => Some("/permalink?share_id=1"),
                    "/permalink" => Some("/login/?dest=%2Fpermalink"),
                    "/straight-to-login" => Some("/login/"),
                    _ => None,
                };
                let response = match location {
                    Some(location) => format!(
                        "HTTP/1.1 301 Moved Permanently\r\nLocation: {}\r\nContent-Length: 0\r\nConnection: close\r\n\r\n",
                        location
                    ),
                    None => "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
                        .to_string(),
                };
                let _ = stream.write_all(response.as_bytes());
            }
        });

        base
    }

    fn local_follower() -> LinkFollowReplacer {
        LinkFollowReplacer::new(
            "local".to_string(),
            Regex::new(r"http://127\.0\.0\.1:\d+/\S+").unwrap(),
            None,
            None,
            None,
            Some(Regex::new(r"/login").unwrap()),
        )
    }

    #[tokio::test]
    async fn stops_before_redirect_to_bad_url() {
        let base = redirect_server();
        let mut replacer = local_follower();

        // reddit sends share links to the permalink, and the permalink to a
        // login page - the permalink is what we want to keep
        assert_eq!(
            replacer.replace(&format!("{}/share", base)).await.unwrap(),
            format!("{}/permalink", base)
        );
        // if the link itself goes straight to the bad url there's nothing
        // worth keeping
        assert_eq!(
            replacer
                .replace(&format!("{}/straight-to-login", base))
                .await
                .unwrap(),
            ""
        );
        // links that don't redirect are returned as is
        assert_eq!(
            replacer
                .replace(&format!("{}/permalink-ok", base))
                .await
                .unwrap(),
            format!("{}/permalink-ok", base)
        );
    }
}

use std::collections::{BTreeMap, HashMap};
use std::ptr::replace;
use std::time::Duration;
use anyhow::anyhow;
use futures::future::try_join_all;
use futures::{stream, StreamExt};
use log::{debug, error, info};
use regex::Regex;
use reqwest::{Error, Response, StatusCode};
use serde_derive::{Deserialize, Serialize};
use serde_with::skip_serializing_none;
use tokio::task::JoinHandle;
use tokio::time::sleep;
use crate::replacer::StringReplacer;

/// Follows links, and then runs an optional replacement at the end
#[skip_serializing_none]
#[derive(Deserialize, Serialize, Debug)]
pub struct LinkFollowReplacer {
    name: String,

    #[serde(with = "serde_regex")]
    match_regex: Regex,

    /// `destination_regex` requires `destination_replacement` be set
    #[serde(with = "serde_regex")]
    #[serde(default)]
    destination_regex: Option<Regex>,

    #[serde(skip_serializing_if = "Option::is_none")]
    destination_replacement: Option<String>,

    /// replace keys with values after link has been followed
    #[serde(skip_serializing_if = "Option::is_none")]
    post_replacement: Option<Vec<Replacements>>,

    #[serde(skip)]
    client: Option<reqwest::Client>,
}

#[derive(Serialize, Deserialize, Debug)]
struct Replacements {
    replace: String,
    with: String,
}

pub fn create_client() -> anyhow::Result<reqwest::Client> {
    let client = reqwest::ClientBuilder::default().user_agent("curl/4.0").build()?;

    Ok(client)
}

impl LinkFollowReplacer {
    pub fn new(
        name: String,
        match_regex: Regex,
        destination_regex: Option<Regex>,
        destination_replacement: Option<String>,
        post_replacement: Option<Vec<Replacements>>,
        client: Option<reqwest::Client>,
    ) -> Self {
        Self {
            name,
            match_regex,
            destination_regex,
            destination_replacement,
            post_replacement,
            client,
        }
    }
}

impl StringReplacer for LinkFollowReplacer {
    fn matches(&self, message: &String) -> bool {
        self.match_regex.is_match(message)
    }

    async fn replace(&mut self, message: &String) -> anyhow::Result<String> {
        if self.client.is_none() {
            self.client = Some(create_client()?);
        }

        if !self.matches(message) {
            return Ok(message.to_string());
        }

        let links: Vec<String> = self.match_regex.find_iter(message.as_str()).map(|link| link.as_str().to_string()).collect();

        let client = &self.client.clone().unwrap();
        let results = futures::stream::iter(
            links.into_iter().map(|link| {
                async move {
                    let mut response = client.get(&link)
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
        ).buffer_unordered(8).collect::<Vec<anyhow::Result<String>>>().await;


        let mut visited_links: Vec<String> = Vec::new();

        for result in results {
            match result {
                Ok(gotten_link) => {
                    let mut link = gotten_link.clone();


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
                }
                Err(err) => {
                    return Err(anyhow!("error visiting link: {}", err));
                }
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
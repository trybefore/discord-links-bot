use regex::Regex;
use serde_derive::{Deserialize, Serialize};
use crate::replacer::StringReplacer;

/// Follows links, and then runs an optional replacement at the end
#[derive(Deserialize, Serialize, Debug)]
pub struct LinkFollowReplacer {
    name: String,

    #[serde(with = "serde_regex")]
    match_regex: Option<Regex>,

    #[serde(with = "serde_regex")]
    destination_regex: Option<Regex>,
    destination_replacement: Option<String>,

    post_replacement: Option<String>,
}

impl LinkFollowReplacer {
    pub fn new(
        name: String,
        match_regex: Option<Regex>,
        destination_regex: Option<Regex>,
        destination_replacement: Option<String>,
        post_replacement: Option<String>,
    ) -> Self {
        Self {
            name,
            match_regex,
            destination_regex,
            destination_replacement,
            post_replacement,
        }
    }
}

impl StringReplacer for LinkFollowReplacer {
    fn matches(&self, message: &String) -> bool {
        todo!()
    }

    fn replace(&self, message: &String) -> anyhow::Result<String> {
        todo!()
    }

    fn name(&self) -> &String {
        &self.name
    }
}
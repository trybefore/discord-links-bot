use crate::replacer;
use regex::Regex;
use serde::{Deserialize, Serialize};
use url::Url;

#[derive(Deserialize, Serialize, Debug)]
pub struct RegexReplacer {
    name: String,
    #[serde(with = "serde_regex")]
    match_regex: Regex,
    replacement: String,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    strip_params: Vec<String>,
}

impl RegexReplacer {
    #[allow(dead_code)]
    pub fn new(name: String, regex: String, replacement: String) -> Result<Self, regex::Error> {
        let rgx = Regex::new(regex.as_str())?;

        Ok(Self {
            match_regex: rgx,
            replacement,
            name,
            strip_params: Vec::new(),
        })
    }
}

/// Remove query parameters listed in `strip_params` from `url_str`, and drop
/// any empty parameters (e.g. a trailing `&` in the replacement string).
fn clean_url(url_str: &str, strip_params: &[String]) -> String {
    let Ok(mut url) = Url::parse(url_str) else {
        return url_str.to_string();
    };

    if url.query().is_none() {
        return url.to_string();
    }

    let params: Vec<(String, String)> = url
        .query_pairs()
        .filter(|(key, _)| !key.is_empty() && !strip_params.iter().any(|p| p == key.as_ref()))
        .map(|(k, v)| (k.into_owned(), v.into_owned()))
        .collect();

    if params.is_empty() {
        url.set_query(None);
    } else {
        let mut query = url.query_pairs_mut();
        query.clear();
        for (k, v) in &params {
            query.append_pair(k, v);
        }
    }

    url.to_string()
}

impl replacer::StringReplacer for RegexReplacer {
    fn matches(&self, message: &str) -> bool {
        self.match_regex.is_match(message)
    }

    async fn replace(&mut self, message: &str) -> anyhow::Result<String> {
        if !self.matches(message) {
            return Ok(message.to_string());
        }
        let result: Vec<String> = self
            .match_regex
            .find_iter(message)
            .map(|link| {
                let replaced = self
                    .match_regex
                    .replace(link.as_str(), &self.replacement)
                    .to_string();
                clean_url(&replaced, &self.strip_params)
            })
            .collect();

        Ok(result.join("\n"))
    }

    fn name(&self) -> &String {
        &self.name
    }
}

use regex::Regex;
use crate::replacer;
use serde::{Deserialize, Serialize};


#[derive(Deserialize, Serialize, Debug)]
pub struct RegexReplacer {
    name: String,
    #[serde(with = "serde_regex")]
    match_regex: Regex,
    replacement: String,
}


// impl RegexReplacer {
//     pub fn new(name: String, regex: String, replacement: String) -> Result<Self, regex::Error> {
//         let rgx = Regex::new(regex.clone().as_str())?;
// 
//         Ok(
//             Self {
//                 match_regex: rgx,
//                 replacement,
//                 name,
//             }
//         )
//     }
// }

impl replacer::StringReplacer for RegexReplacer {
    fn matches(&self, message: &String) -> bool {
        self.match_regex.is_match(message.as_str())
    }

    async fn replace(&mut self, message: &String) -> anyhow::Result<String> {
        if !self.matches(message) {
            return Ok(message.to_string());
        }
        let result: Vec<String> = self.match_regex.find_iter(message.as_str()).map(|link| {
            self.match_regex.replace(link.as_str(), &self.replacement).to_string()
        }).collect();

        Ok(result.join("\n"))
    }

    fn name(&self) -> &String {
        &self.name
    }
}

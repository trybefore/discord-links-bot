mod replacer_regex;
mod replacer_link_follower;


use std::sync::Arc;
use std::time::Duration;
use anyhow::{anyhow, bail};
use config::Config;
use log::{debug};
use serde::{Deserialize};
use serde_derive::Serialize;
use thiserror::Error;
use tokio::time::sleep;
use crate::replacer::replacer_link_follower::LinkFollowReplacer;
use crate::replacer::replacer_regex::RegexReplacer;


pub async fn replace_message(message: &String, config: Arc<Config>) -> anyhow::Result<String> {
    let replacers = get_matching_replacers(message, config)?;

    let mut links: Vec<String> = Vec::new();

    for mut r in replacers {
        debug!("replacing {} with replacer {}", message, r.name());
        let mut replaced_link = r.replace(&message).await?;
        if message.contains("||") {
            replaced_link = format!("||{}||", replaced_link);
        }
        links.push(replaced_link);
    }

    Ok(links.join("\n"))
}


pub fn get_matching_replacers(message: &String, config: Arc<Config>) -> anyhow::Result<Vec<Replacer>> {
    debug!("checking message: {}", message);
    let mut replacers: Vec<Replacer> = get_replacers(config)?.0.into_iter().filter(|r| r.matches(message)).collect();

    /*
    return if !r.matches(message) {
        debug!("replacer {} does not match the message", r.name());
        false
    } else { true }
     */

    if replacers.len() == 0 {
        debug!("found no replacers by that message");
        bail!(ReplacerError::NoReplacerFound)
    }

    for x in &replacers {
        debug!("found replacer: {}", x.name());
    }


    replacers.sort_by_key(|r| r.name().clone()); // sort alphabetically, for consistent test results
    Ok(replacers)
}

/// get all replacers in config.yaml
fn get_replacers(config: Arc<Config>) -> anyhow::Result<Replacers> {
    //let replacers = crate::config::SETTINGS.read().unwrap().get::<Replacers>("replacers")?;

    let replacers = config.get::<Replacers>("replacers")?;

    Ok(replacers)
}

/// get replacer by name, if it exists
pub fn get_replacer_by_name(name: &String, config: Arc<Config>) -> anyhow::Result<Option<Replacer>> {
    let replacers = get_replacers(config)?;
    let result = replacers.0.into_iter().filter_map(|r| {
        if r.name().eq(name) { Some(r) } else { None }
    }).last();


    Ok(
        result
    )
}

/// get all tests in config.yaml
pub fn get_tests(config: Arc<Config>) -> anyhow::Result<Tests> {
    Ok(
        config.get::<Tests>("tests")?
    )
}

/// get tests for replacer, if it has any
pub fn get_tests_by_name(replacer_name: String, config: Arc<Config>) -> anyhow::Result<Tests> {
    let tests = get_tests(config)?.filter_by_replacer_name(replacer_name);


    Ok(tests)
}

/// run tests for replacer_name if provided, all replacers if `None`
pub async fn run_replacer_tests(replacer_name: Option<String>, config: Arc<Config>) -> anyhow::Result<()> {
    let replacer_tests: Tests;

    if let Some(name) = replacer_name {
        replacer_tests = get_tests_by_name(name, config.clone())?;
    } else {
        replacer_tests = get_tests(config.clone())?;
    }

    run_tests(replacer_tests, config).await?;

    Ok(())
}

async fn run_tests(replacer_tests: Tests, config: Arc<Config>) -> anyhow::Result<()> {
    debug!("running tests for {} replacers", replacer_tests.0.len());
    for replacer in replacer_tests.0 {
        debug!("running {} tests for replacer {}", replacer.replacer_name, replacer.tests.len());
        let name = &replacer.replacer_name;
        let tests = &replacer.tests;

        if let Some(mut replacer) = get_replacer_by_name(&name, config.clone())? {
            let mut test_count = 0;
            for test in tests {
                match replacer {
                    Replacer::LinkReplacer(_) => {
                        if cfg!(feature = "skip-link-followers") {
                            continue;
                        }
                    }
                    _ => {}
                }
                test_count += 1;
                let got = replacer.replace(&test.have).await.unwrap_or_else(|_| "".to_string());
                let want = test.want.clone();

                if !got.eq(&want) {
                    return Err(anyhow!("{} #{}: {} != {}", &name, &test_count, got, want));
                }
                //debug!("have: {}, want: {}, got: {}", test.have, want, got);
                sleep(Duration::from_millis(500)).await; // sleep to avoid TooManyRequests for link followers
            }
        } else {
            return Err(anyhow!("no replacer found by name {}", &name));
        }
    }

    Ok(())
}


#[derive(Error, Debug)]
pub enum ReplacerError {
    #[error("Could not find any matching replacers")]
    NoReplacerFound
}

#[derive(Serialize, Deserialize)]
pub struct Tests(pub(crate) Vec<ReplacerTests>);

impl Tests {
    fn filter_by_replacer_name(&self, name: String) -> Tests {
        let old_tests = self.0.clone();

        let tests = old_tests.into_iter().filter_map(|r| {
            if r.replacer_name.eq(&name) { Some(r) } else { None }
        }).collect::<Vec<ReplacerTests>>();

        let result = Tests(tests);


        result
    }
}

impl FromIterator<ReplacerTests> for Tests {
    fn from_iter<T: IntoIterator<Item=ReplacerTests>>(iter: T)
                                                      -> Self {
        let mut c = Vec::new();

        for tests in iter {
            c.push(tests.clone());
        }

        Self { 0: c }
    }
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ReplacerTests {
    replacer_name: String,
    tests: Vec<ReplacerTest>,
}

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct ReplacerTest {
    have: String,
    want: String,
}

#[derive(Debug, Deserialize, Serialize)]
#[serde(tag = "type")]
pub enum Replacer {
    #[serde(rename = "regex")]
    Regex(RegexReplacer),
    #[serde(rename = "link-follower")]
    LinkReplacer(LinkFollowReplacer),
}

impl StringReplacer for Replacer {
    fn matches(&self, message: &String) -> bool {
        match self {
            Replacer::Regex(r) => r.matches(message),
            Replacer::LinkReplacer(r) => r.matches(message)
        }
    }

    async fn replace(&mut self, message: &String) -> anyhow::Result<String> {
        match self {
            Replacer::Regex(r) => r.replace(message).await,
            Replacer::LinkReplacer(r) => r.replace(message).await
        }
    }

    fn name(&self) -> &String {
        match self {
            Replacer::Regex(r) => r.name(),
            Replacer::LinkReplacer(r) => r.name()
        }
    }
}

#[derive(Deserialize, Serialize, Debug)]
pub struct Replacers(pub(crate) Vec<Replacer>);

pub trait StringReplacer {
    /// check if replacer has any matches for message
    fn matches(&self, message: &String) -> bool;
    /// replace occurrences in message, if applicable
    async fn replace(&mut self, message: &String) -> anyhow::Result<String>;

    /// name of the replacer
    fn name(&self) -> &String;
}


#[cfg(test)]
mod tests {
    use std::sync::Arc;
    use config::Config;
    use log::{debug, info};
    use log::LevelFilter::Debug;

    use crate::replacer::replacer_regex::RegexReplacer;
    use crate::replacer::{Replacers, Replacer, run_replacer_tests, get_matching_replacers, StringReplacer, replace_message};

    #[test]
    fn serialize_test() {
        let replacer = RegexReplacer::new("twitter".to_string(), r"https?://(?P<tld>twitter|x)\.com/(?:#!/)?(\w+)/status(es)?/(\d+)".to_string(), "https://vxtwitter.com/$2/status/$4".to_string()).unwrap();
        let replacers = Replacers(vec![Replacer::Regex(replacer)]);


        println!("{:?}", replacers);
    }

    #[test]
    fn deserialize_test() {
        let replacer = RegexReplacer::new("twitter".to_string(), r"https?://(?P<tld>twitter|x)\.com/(?:#!/)?(\w+)/status(es)?/(\d+)".to_string(), "https://vxtwitter.com/$2/status/$4".to_string()).unwrap();
        let replacers = Replacers(vec![Replacer::Regex(replacer)]);

        let want_yaml = r"- type: regex
  name: twitter
  match_regex: https?://(?P<tld>twitter|x)\.com/(?:#!/)?(\w+)/status(es)?/(\d+)
  replacement: https://vxtwitter.com/$2/status/$4";
        let got_yaml = serde_yaml::to_string(&replacers).unwrap();

        _ = serde_yaml::from_str::<Replacers>(want_yaml).unwrap();

        assert_eq!(want_yaml, got_yaml.trim_end())
    }

    fn create_config() -> Arc<Config> {
        Arc::new(crate::config::create().unwrap())
    }

    #[tokio::test]
    async fn run_file_tests() {
        let config = create_config();
        let result = run_replacer_tests(None, config).await;
        assert!(result.is_ok(), "failed to run tests: {:?}", result.err().unwrap())
    }

    #[tokio::test]
    async fn test_replace_message() {
        let config = create_config();
        _ = env_logger::builder().is_test(true).filter_level(Debug).try_init();
        let message = r#"https://www.tiktok.com/@realcompmemer/video/7314546788617309471
https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4"#.to_string();

        let new_message = replace_message(&message, config).await.unwrap();

        let want = r"https://cdn.discordapp.com/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4
https://www.vxtiktok.com/@realcompmemer/video/7314546788617309471";
        info!("got new message: {}", new_message);

        assert_eq!(new_message, want);
    }

    #[test]
    fn test_get_matching_replacers() {
        let config = create_config();
        env_logger::builder().is_test(true).filter_level(Debug).init();
        let message = r#"https://www.tiktok.com/@realcompmemer/video/7314546788617309471
https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4 "#.to_string();
        let matching_replacers = get_matching_replacers(&message, config).expect("could not read replacers from config");

        assert!(matching_replacers.len() > 0);

        let discord_name = matching_replacers.get(0).unwrap().name();
        let tiktok_name = matching_replacers.get(1).unwrap().name();

        debug!("discord_name = {} tiktok_name = {}", discord_name, tiktok_name);

        assert!(matching_replacers.get(0).unwrap().name().eq("discord"));
        assert!(matching_replacers.get(1).unwrap().name().eq("tiktok"))
    }
}
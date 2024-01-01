mod replacer_regex;
mod replacer_links_follower;


use std::fmt::Display;
use anyhow::anyhow;
use serde::{Deserialize, Deserializer};
use serde_derive::Serialize;
use crate::config::SETTINGS;
use crate::replacer::replacer_links_follower::LinkFollowReplacer;
use crate::replacer::replacer_regex::RegexReplacer;


/// get all replacers in config.yaml
fn get_replacers() -> anyhow::Result<Replacers> {
    let replacers = crate::config::SETTINGS.read().unwrap().get::<Replacers>("replacers")?;


    Ok(replacers)
}

/// get replacer by name, if it exists
pub fn get_replacer_by_name(name: &String) -> anyhow::Result<Option<Replacer>> {
    let replacers = get_replacers()?;
    let result = replacers.0.into_iter().filter_map(|r| {
        if r.name().eq(name) { Some(r) } else { None }
    }).last();


    Ok(
        result
    )
}

/// get all tests in config.yaml
pub fn get_tests() -> anyhow::Result<Tests> {
    Ok(
        SETTINGS.read().unwrap().get::<Tests>("tests")?
    )
}

/// get tests for replacer, if it has any
pub fn get_tests_by_name(replacer_name: String) -> anyhow::Result<Tests> {
    let tests = get_tests()?.filter_by_replacer_name(replacer_name);


    Ok(tests)
}

pub fn run_replacer_tests(replacer_name: Option<String>) -> anyhow::Result<()> {
    let mut replacer_tests: Tests = Tests(vec![]);

    if let Some(name) = replacer_name {
        replacer_tests = get_tests_by_name(name)?;
    } else {
        replacer_tests = get_tests()?;
    }

    run_tests(replacer_tests)?;

    Ok(())
}

fn run_tests(replacer_tests: Tests) -> anyhow::Result<()> {
    for replacer in replacer_tests.0 {
        let name = &replacer.replacer_name;
        let tests = &replacer.tests;

        if let Some(replacer) = get_replacer_by_name(&name)? {
            for test in tests {
                let got = replacer.replace(&test.have)?;
                let want = test.want.clone();

                if !got.eq(&want) {
                    return Err(anyhow!("{} != {}", got, want));
                }
            }
        } else {
            return Err(anyhow!("no replacer found by name {}", &name));
        }
    }

    Ok(())
}


#[derive(Serialize, Deserialize)]
pub struct Tests(pub(crate) Vec<ReplacerTests>);

impl Tests {
    fn filter_by_replacer_name(&self, name: String) -> Tests {
        let old_tests = self.0.clone();

        let tests = old_tests.into_iter().filter_map(|r| {
            if r.replacer_name.eq(&name) { Some(r) } else { None }
        }).collect::<Vec::<ReplacerTests>>();

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

    fn replace(&self, message: &String) -> anyhow::Result<String> {
        match self {
            Replacer::Regex(r) => r.replace(message),
            Replacer::LinkReplacer(r) => r.replace(message)
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
    fn replace(&self, message: &String) -> anyhow::Result<String>;

    /// name of the replacer
    fn name(&self) -> &String;
}


#[cfg(test)]
mod tests {
    use crate::replacer::replacer_regex::RegexReplacer;
    use crate::replacer::{Replacers, Replacer, run_replacer_tests};

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
        //println!("got: {}", &got_yaml);

        _ = serde_yaml::from_str::<Replacers>(want_yaml).unwrap();

        assert_eq!(want_yaml, got_yaml.trim_end())
    }

    #[test]
    fn run_file_tests() {
        assert!(run_replacer_tests(None).is_ok())
    }
}
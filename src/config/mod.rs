
use std::path::Path;
use std::sync::mpsc::channel;
use std::sync::RwLock;
use std::time::Duration;
use config::{Config, File};
use lazy_static::lazy_static;
use notify::{Event, RecommendedWatcher, RecursiveMode, Watcher};

lazy_static! {
    pub static ref SETTINGS: RwLock<Config> = RwLock::new(
        {
            let settings = Config::builder().add_source(File::with_name(&CONFIG_PATH));


            settings.build().unwrap()
        }
    );

    pub static ref CONFIG_PATH: String = {
        match std::env::var("BOT_CONFIG_FILE") {
            Ok(result) => {result},
            Err(_) => {"./config.yaml".to_string()}
        }
    };
}



pub fn start_watcher() {
    let (tx, rx) = channel();
    let mut watcher: RecommendedWatcher = Watcher::new(tx, notify::Config::default().with_poll_interval(Duration::from_secs(2))).unwrap();
    watcher.watch(Path::new(&CONFIG_PATH.clone().to_string()), RecursiveMode::NonRecursive).unwrap();
    loop {
        #[allow(deprecated)] match rx.recv() {
            Ok(Ok(Event { kind: notify::event::EventKind::Modify(_), .. })) => {
                println!("{} updated, refreshing config", &CONFIG_PATH.clone().to_string());
                SETTINGS.write().unwrap().refresh().unwrap();

                print_config()
            }
            Err(_) => {}
            _ => {}
        }
    }
}

pub fn print_config() {}
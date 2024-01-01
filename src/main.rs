use std::thread;
use clap::{Parser, Subcommand};
use env_logger::Env;
use log::{error, info, Level, log_enabled};
use crate::replacer::run_replacer_tests;

mod replacer;
mod discordbot;
mod config;


#[derive(Debug, Parser)]
#[command(name = "linksbot")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Debug, Subcommand)]
enum Commands {
    Test {
        /// Replacer name, leave empty to run tests on all replacers
        replacer_name: Option<String>,
    },
    Run {},
}

fn start() {
    let handle = thread::spawn(|| {
        config::start_watcher();
    });


    handle.join().unwrap();
}


#[tokio::main]
async fn main() -> anyhow::Result<()> {
    env_logger::init();

    let args = Cli::parse();

    match args.command {
        Commands::Test { replacer_name } => {
            match run_replacer_tests(replacer_name).await {
                Ok(_) => {
                    info!("successfully ran tests")
                }
                Err(err) => {
                    error!("failed to run tests: {}", err)
                }
            };
        }
        Commands::Run { .. } => {}
    }

    Ok(())
}


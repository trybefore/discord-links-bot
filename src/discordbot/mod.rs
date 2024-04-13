mod event_handler;


use std::sync::Arc;
use anyhow::{bail};
use config::Config;
use serenity::all::{GatewayIntents};
use serenity::prelude::*;
use crate::discordbot::event_handler::Handler;

pub async fn create_client(config: Arc<Config>) -> anyhow::Result<Client> {
    let token = match std::env::var("BOT_TOKEN") {
        Ok(token) => token,
        Err(err) => bail!("failed to read BOT_TOKEN environment variable: {}", err)
    };

    let intents = GatewayIntents::GUILDS | GatewayIntents::GUILD_MESSAGES | GatewayIntents::MESSAGE_CONTENT | GatewayIntents::GUILD_PRESENCES;

    let handler = Handler::new(config);
    let client = Client::builder(token, intents).event_handler(handler).await?;

    Ok(client)
}
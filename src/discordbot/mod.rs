mod event_handler;


use anyhow::{bail};
use serenity::all::{GatewayIntents};
use serenity::prelude::*;
use crate::discordbot::event_handler::Handler;

pub async fn create_client() -> anyhow::Result<Client> {
    let token = match std::env::var("BOT_TOKEN") {
        Ok(token) => token,
        Err(err) => bail!("failed to read BOT_TOKEN environment variable: {}", err)
    };

    let intents = GatewayIntents::GUILDS | GatewayIntents::GUILD_MESSAGES | GatewayIntents::MESSAGE_CONTENT | GatewayIntents::GUILD_PRESENCES;

    let client = Client::builder(token, intents).event_handler(Handler).await?;

    Ok(client)
}
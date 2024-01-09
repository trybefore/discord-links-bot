mod event_handler;

use serenity::all::{GatewayIntents};
use serenity::prelude::*;
use crate::discordbot::event_handler::Handler;

pub async fn create_client() -> anyhow::Result<Client> {
    let token = std::env::var("BOT_TOKEN")?;

    let intents = GatewayIntents::GUILDS | GatewayIntents::GUILD_MESSAGES | GatewayIntents::MESSAGE_CONTENT | GatewayIntents::GUILD_PRESENCES;

    let client = Client::builder(token, intents).event_handler(Handler).await?;

    Ok(client)
}
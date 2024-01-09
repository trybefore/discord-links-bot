use log::info;
use serenity::all::{Message, Ready};
use serenity::async_trait;
use serenity::prelude::*;
use serenity::gateway::ActivityData;
use serenity::model::user::OnlineStatus;

pub(crate) struct Handler;

#[async_trait]
impl EventHandler for Handler {
    async fn message(&self, _ctx: Context, _msg: Message) {
        // info!("[{}] [{}] {}: {}", msg.guild_id.unwrap(), msg.channel_id, msg.author.name, msg.content);
        
        
    }

    async fn ready(&self, ctx: Context, data: Ready) {
        info!("now live in {} guilds",data.guilds.len());


        let activity = ActivityData::competing(format!("{}", std::env::var("GIT_HASH").unwrap_or("dunno".to_string())));
        let status = OnlineStatus::DoNotDisturb;

        ctx.set_presence(Some(activity), status)
    }
}
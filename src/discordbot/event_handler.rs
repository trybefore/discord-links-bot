use std::time::Duration;
use anyhow::{bail};
use futures::StreamExt;
use log::{debug, error, info};
use serenity::all::{CreateAttachment, CreateMessage, EditMessage, Event, Message, Ready};
use serenity::async_trait;
use serenity::builder::CreateAllowedMentions;
use serenity::prelude::*;
use serenity::gateway::ActivityData;
use serenity::model::user::OnlineStatus;
use tokio::time::Instant;
use crate::replacer::{ReplacerError};
use crate::resource::{GUH, NOREG};


pub(crate) struct Handler;


pub async fn replace_discord_message(ctx: &Context, msg: &Message) -> anyhow::Result<()> {
    debug!("[{}] [{}] {}: {}", msg.guild_id.unwrap(), msg.channel_id, msg.author.name, &msg.content);

    let new_message = match crate::replacer::replace_message(&msg.content).await {
        Ok(msg) => msg,
        Err(err) => {
            match err.downcast_ref() {
                Some(ReplacerError::NoReplacerFound) => {
                    debug!("no replacer found for that message, ignoring it");
                    return Ok(());
                }
                _ => { bail!("unexpected error: {}", err) }
            }
        }
    };

    let response = CreateMessage::new().content(&new_message).reference_message(msg).allowed_mentions(CreateAllowedMentions::new().replied_user(false));
    debug!("replaced [{}] -> [{}]", &msg.content, &new_message);


    let start_time = Instant::now();
    match msg.channel_id.send_message(&ctx.http, response).await {
        Ok(_) => {
            debug!("took {}ms to send message", start_time.elapsed().as_millis());

            hide_embeds(&ctx, &mut msg.clone()).await?;

            return Ok(());
        }
        Err(err) => bail!(err)
    }
}

async fn hide_embeds(ctx: &Context, msg: &mut Message) -> anyhow::Result<()> {
    let msg_id = msg.id;

    let mut message_updates = serenity::collector::collect(&ctx.shard, move |event: &Event| {
        match event {
            Event::MessageUpdate(updated_message) if updated_message.id == msg_id => Some(updated_message.id),
            _ => None,
        }
    });


    let _ = tokio::time::timeout(Duration::from_millis(2000), message_updates.next()).await;
    msg.edit(&ctx, EditMessage::new().suppress_embeds(true)).await?;


    Ok(())
}

#[async_trait]
impl EventHandler for Handler {
    async fn message(&self, ctx: Context, msg: Message) {
        if ctx.http.get_current_user().await.unwrap().id.to_string().eq(&msg.author.id.to_string()) {
            debug!("ignoring message, as it is from the bot");
            return;
        }
        let match_message = msg.content.to_lowercase();

        if match_message.contains("guh") || match_message.contains("norway") {
            let mut response = CreateMessage::new();

            if msg.content.to_lowercase().contains("guh") {
                response = response.add_file(CreateAttachment::bytes(GUH.to_vec(), "guh.gif"));
            }

            if msg.content.to_lowercase().contains("norway") {
                response = response.add_file(CreateAttachment::bytes(NOREG.to_vec(), "norway.png"));
            }

            response = response.reference_message(&msg).allowed_mentions(CreateAllowedMentions::new().replied_user(false));

            match msg.channel_id.send_message(&ctx.http, response).await {
                Ok(m) => debug!("responded with message {}", m.id),
                Err(err) => error!("error replying to message {}", err),
            };
        }

        if let Err(err) = replace_discord_message(&ctx, &msg).await {
            error!("failed to replace message: {}", err);
            return;
        }
    }


    async fn ready(&self, ctx: Context, data: Ready) {
        info!("now live in {} guilds",data.guilds.len());

        let version = option_env!("GIT_HASH").unwrap_or("dunno");

        let activity = ActivityData::competing(format!("{}", version));
        let status = OnlineStatus::DoNotDisturb;

        ctx.set_presence(Some(activity), status)
    }
}
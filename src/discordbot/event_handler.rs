use crate::replacer::ReplacerError;
use crate::resource::{GUH, NOREG, SCOTLAND};
use anyhow::bail;
use config::Config;
use futures::StreamExt;
use log::{debug, error, info};
use serenity::all::{CreateAttachment, CreateMessage, EditMessage, Event, Message, Ready};
use serenity::async_trait;
use serenity::builder::CreateAllowedMentions;
use serenity::gateway::ActivityData;
use serenity::model::user::OnlineStatus;
use serenity::prelude::*;
use std::sync::Arc;
use std::time::Duration;
use tokio::time::Instant;

pub(crate) struct Handler {
    config: Arc<Config>,
}

impl Handler {
    pub fn new(config: Arc<Config>) -> Self {
        Self { config }
    }

    pub async fn replace_discord_message(
        &self,
        ctx: &Context,
        msg: &Message,
    ) -> anyhow::Result<()> {
        debug!(
            "[{}] [{}] {}: {}",
            msg.guild_id.unwrap(),
            msg.channel_id,
            msg.author.name,
            &msg.content
        );
        let start_time = Instant::now();

        let new_message =
            match crate::replacer::replace_message(&msg.content, self.config.clone()).await {
                Ok(msg) => msg,
                Err(err) => match err.downcast_ref() {
                    Some(ReplacerError::NoReplacerFound) => {
                        debug!("no replacer found for that message, ignoring it");
                        return Ok(());
                    }
                    _ => {
                        bail!("unexpected error: {}", err)
                    }
                },
            };

        if new_message.is_empty() {
            debug!("empty response message (should be okay if the link matched a bad_url_regex)");
            return Ok(());
        }

        let response = CreateMessage::new()
            .content(&new_message)
            .reference_message(msg)
            .allowed_mentions(CreateAllowedMentions::new().replied_user(false));
        debug!("replaced [{}] -> [{}]", &msg.content, &new_message);

        match msg.channel_id.send_message(&ctx.http, response).await {
            Ok(_) => {
                debug!(
                    "took {}ms to visit the links and send the message",
                    start_time.elapsed().as_millis()
                );

                hide_embeds(ctx, &mut msg.clone()).await?;

                Ok(())
            }
            Err(err) => bail!(err),
        }
    }
}

/// Returns true if `keyword` appears in `content` in at least one spot where it
/// isn't immediately followed by an exclamation mark. An occurrence like `guh!`
/// is treated as an opt-out and won't trigger a response on its own.
fn mentions(content: &str, keyword: &str) -> bool {
    content
        .match_indices(keyword)
        .any(|(idx, _)| !content[idx + keyword.len()..].starts_with('!'))
}

async fn hide_embeds(ctx: &Context, msg: &mut Message) -> anyhow::Result<()> {
    let msg_id = msg.id;

    let mut message_updates =
        serenity::collector::collect(&ctx.shard, move |event: &Event| match event {
            Event::MessageUpdate(updated_message) if updated_message.id == msg_id => {
                Some(updated_message.id)
            }
            _ => None,
        });

    let _ = tokio::time::timeout(Duration::from_millis(2000), message_updates.next()).await;
    msg.edit(&ctx, EditMessage::new().suppress_embeds(true))
        .await?;

    Ok(())
}

#[async_trait]
impl EventHandler for Handler {
    async fn message(&self, ctx: Context, msg: Message) {
        if ctx
            .http
            .get_current_user()
            .await
            .unwrap()
            .id
            .to_string()
            .eq(&msg.author.id.to_string())
        {
            debug!("ignoring message, as it is from the bot");
            return;
        }
        let match_message = msg.content.to_lowercase();

        let mentions_guh = mentions(&match_message, "guh");
        let mentions_norway = mentions(&match_message, "norway");
        let mentions_scotland = mentions(&match_message, "scotland");

        if mentions_guh || mentions_norway || mentions_scotland {
            let mut response = CreateMessage::new();

            if mentions_guh {
                response = response.add_file(CreateAttachment::bytes(GUH.to_vec(), "guh.gif"));
            }

            if mentions_norway {
                response = response.add_file(CreateAttachment::bytes(NOREG.to_vec(), "norway.png"));
            }

            if mentions_scotland {
                response =
                    response.add_file(CreateAttachment::bytes(SCOTLAND.to_vec(), "scotland.jpg"));
            }

            response = response
                .reference_message(&msg)
                .allowed_mentions(CreateAllowedMentions::new().replied_user(false));

            match msg.channel_id.send_message(&ctx.http, response).await {
                Ok(m) => debug!("responded with message {}", m.id),
                Err(err) => error!("error replying to message {}", err),
            };
        }

        if let Err(err) = self.replace_discord_message(&ctx, &msg).await {
            error!("failed to replace message: {}", err);
            return;
        }
    }

    async fn ready(&self, ctx: Context, data: Ready) {
        info!("now live in {} guilds", data.guilds.len());

        let version = option_env!("GIT_HASH").unwrap_or("dunno");

        let activity = ActivityData::competing(version.to_string());
        let status = OnlineStatus::DoNotDisturb;

        ctx.set_presence(Some(activity), status)
    }
}

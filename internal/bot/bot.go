package bot

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/spf13/viper"
	"github.com/trybefore/linksbot/internal/config"
	"github.com/trybefore/linksbot/internal/replacer"
)

var botState *state.State
var messageQueue chan replacer.Message

func Run(ctx context.Context) error {
	botToken := viper.GetString(config.BotToken)

	if botToken == "" {
		log.Fatalf("bot token missing in environment variables, flags and config file")
	}
	if !strings.HasPrefix(botToken, "Bot ") {
		botToken = "Bot " + botToken
	}

	messageQueue = make(chan replacer.Message)

	identifier := gateway.DefaultIdentifier(botToken)

	identifier.Presence = &gateway.UpdatePresenceCommand{
		Since: discord.UnixMsTimestamp(time.Now().UnixMilli()),
		Activities: []discord.Activity{
			{
				Type: discord.CompetingActivity,
				Name: config.Commit,
			},
		},
		Status: discord.OnlineStatus,
		AFK:    false,
	}

	botState = state.NewWithIdentifier(identifier)

	botState.AddHandler(func(c *gateway.MessageCreateEvent) {
		me, err := botState.Me()
		if err != nil {
			log.Printf("error finding myself: %v", err)
			return
		}

		if me.ID == c.Author.ID {
			return // ignore own messages
		}

		if replacers := replacer.FindReplacers(c.Message); len(replacers) > 0 {
			messageQueue <- replacer.Message{
				Content:   c.Message,
				Author:    c.Author,
				Replacers: &replacers,
			}
			return
		}
		msg := strings.ToLower(c.Message.Content)
		msgContent := ""
		if viper.GetBool(config.NorwayMentioned) && strings.Contains(msg, "norway") {
			msgContent += "https://i.imgur.com/5msLOh4.png\n"
		} else if viper.GetBool(config.Guh) && strings.Contains(msg, "guh") {
			msgContent += "https://i.imgur.com/mJIDoiI.gif\n"
		}

		if msgContent != "" {
			if _, err := botState.SendMessageComplex(c.Message.ChannelID, api.SendMessageData{
				Content:         msgContent,
				AllowedMentions: &api.AllowedMentions{},
				Reference: &discord.MessageReference{
					MessageID: c.Message.ID,
				},
			}); err != nil {
				log.Println("error sending reply message:", err)
			}
		}

	})

	botState.AddIntents(gateway.IntentGuilds)
	botState.AddIntents(gateway.IntentDirectMessages)
	botState.AddIntents(gateway.IntentGuildMessages)

	if err := botState.Open(ctx); err != nil {
		return err
	}

	log.Printf("discord-links-bot running -- commit: %s", config.Commit)

	defer botState.Close()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	go func(s *state.State) {
		for msg := range messageQueue {
			go replacer.ReplaceMessage(s, msg)
		}
	}(botState)

	<-c
	log.Printf("terminating signal received, stopping bot")
	close(messageQueue)

	return context.Canceled
}

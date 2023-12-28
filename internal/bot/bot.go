package bot

import (
	"bytes"
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
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/spf13/viper"
	"github.com/trybefore/linksbot/internal/config"
	"github.com/trybefore/linksbot/internal/replacer"
	"github.com/trybefore/linksbot/resource"
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
			log.Printf("error: finding myself: %v", err)
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

		var files []sendpart.File

		if viper.GetBool(config.NorwayMentioned) && strings.Contains(msg, "norway") {
			file, err := resource.FS.ReadFile("resources/norway.png")
			if err == nil {
				files = append(files, sendpart.File{
					Name:   "norway.png",
					Reader: bytes.NewReader(file),
				})
			} else {
				log.Println("error: reading file:", err)
			}
		}

		if viper.GetBool(config.Guh) && strings.Contains(msg, "guh") {
			file, err := resource.FS.ReadFile("resources/guh.gif")
			if err == nil {
				files = append(files, sendpart.File{
					Name:   "guh.gif",
					Reader: bytes.NewReader(file),
				})
			} else {
				log.Println("error: reading file:", err)
			}
		}

		if msgContent != "" && len(files) == 0 {
			if _, err := botState.SendMessageComplex(c.Message.ChannelID, api.SendMessageData{
				Content:         msgContent,
				AllowedMentions: &api.AllowedMentions{},
				Reference: &discord.MessageReference{
					MessageID: c.Message.ID,
				},
			}); err != nil {
				log.Println("error: sending reply message:", err)
			}
		} else if len(files) > 0 {
			if _, err := botState.SendMessageComplex(c.Message.ChannelID, api.SendMessageData{
				AllowedMentions: &api.AllowedMentions{},
				Reference: &discord.MessageReference{
					MessageID: c.Message.ID,
				},

				Files: files,
			}); err != nil {
				log.Println("error: sending file replies:", err)
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
			go replacer.SendReplacementMessage(s, msg)
		}
	}(botState)

	<-c
	log.Printf("terminating signal received, stopping bot")
	close(messageQueue)

	return context.Canceled
}

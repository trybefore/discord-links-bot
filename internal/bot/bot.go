package bot

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/spf13/viper"
	"github.com/trybefore/linksbot/internal/config"
	"github.com/trybefore/linksbot/internal/debug"
	"github.com/trybefore/linksbot/internal/replacer"
)

var botSession *session.Session
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

	botSession = session.New(botToken)

	botSession.AddHandler(func(c *gateway.MessageCreateEvent) {
		me, err := botSession.Me()
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
		if viper.GetBool(config.NorwayMentioned) && strings.Contains(strings.ToLower(c.Message.Content), "norway") {
			if _, err := botSession.SendMessageComplex(c.Message.ChannelID, api.SendMessageData{
				Content:         "https://cdn.discordapp.com/attachments/735399993485033472/1071891567389974608/x5Lqn.png",
				AllowedMentions: &api.AllowedMentions{},
				Reference: &discord.MessageReference{
					MessageID: c.Message.ID,
				},
			}); err != nil {
				log.Println("error sending reply message:", err)
			}
		}
	})

	botSession.AddIntents(gateway.IntentDirectMessages)
	botSession.AddIntents(gateway.IntentGuildMessages)

	if err := botSession.Open(ctx); err != nil {
		return err
	}

	log.Printf("discord-links-bot running -- commit: %s", debug.Commit)

	defer botSession.Close()

	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt)

	go func(s *session.Session) {
		for msg := range messageQueue {
			go replacer.ReplaceMessage(s, msg)
		}
	}(botSession)

	<-c
	log.Printf("terminating signal received, stopping bot")
	close(messageQueue)
	return nil // TODO: implement
}

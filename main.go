package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	var token = os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Fatalln("$BOT_TOKEN missing")
	}

	s := session.New("Bot " + token)

	queue := make(chan message)

	s.AddHandler(func(c *gateway.MessageCreateEvent) {
		if strings.HasPrefix(c.Message.Content, "https://twitter.com") {
			queue <- message{
				content: c.Message,
				author:  c.Author,
			}
		}
	})

	s.AddIntents(gateway.IntentDirectMessages)
	s.AddIntents(gateway.IntentGuildMessages)

	if err := s.Open(context.Background()); err != nil {
		log.Fatalln("Failed to connect:", err)
	}

	defer s.Close()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func(s *session.Session) {
		for msg := range queue {
			go replaceMessage(s, msg)
		}
	}(s)

	<-c
	close(queue)
}

func replaceMessage(s *session.Session, m message) {

	newMessage := fmt.Sprintf(`from: %s
	%s
	`, m.author.Mention(), strings.Replace(m.content.Content, "https://twitter.com", "https://fxtwitter.com", 1))

	_, err := s.SendMessage(m.content.ChannelID, newMessage)
	if err != nil {
		log.Println("error sending message:", err)
	}
	if err = s.DeleteMessage(m.content.ChannelID, m.content.ID, api.AuditLogReason("fxtwitterbot")); err != nil {
		log.Println("error deleting message:", err)
	}
}

type message struct {
	content discord.Message
	author  discord.User
}

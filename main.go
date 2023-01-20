package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/joho/godotenv"
)

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}()

func main() {
	_ = godotenv.Load()

	log.Printf("discord-links-bot commit %s", Commit)

	var token = os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Fatalln("$BOT_TOKEN missing")
	}

	s := session.New("Bot " + token)

	queue := make(chan message)

	s.AddHandler(func(c *gateway.MessageCreateEvent) {
		me, err := s.Me()
		if err != nil {
			log.Printf("error finding myself: %v", err)
		}
		if me.ID == c.Author.ID {
			return
		}
		if replacers := findReplacers(c.Message); len(replacers) > 0 {
			queue <- message{
				content:   c.Message,
				author:    c.Author,
				replacers: &replacers,
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

	go startHealthCheck()

	go func(s *session.Session) {
		for msg := range queue {
			go replaceMessage(s, msg)
		}
	}(s)

	<-c
	close(queue)
}

func startHealthCheck() {
	http.HandleFunc("/health_check", getHealth)
	if err := http.ListenAndServe(":"+os.Getenv("HEALTHCHECK_PORT"), nil); err != nil {
		log.Fatal(err)
	}
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type message struct {
	content   discord.Message
	author    discord.User
	replacers *[]replacer
}

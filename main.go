package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/joho/godotenv"
)

var twitterRegex = regexp.MustCompile(`https?:\/\/(?P<tld>twitter)\.com\/(?:#!\/)?(\w+)\/status(es)?\/(\d+)`)

func main() {
	_ = godotenv.Load()

	var token = os.Getenv("BOT_TOKEN")

	if token == "" {
		log.Fatalln("$BOT_TOKEN missing")
	}

	s := session.New("Bot " + token)

	queue := make(chan message)

	s.AddHandler(func(c *gateway.MessageCreateEvent) {
		if twitterRegex.MatchString(c.Content) {
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

func createOutput(m message) string {
	matches := twitterRegex.FindAllString(m.content.Content, -1)

	if matches == nil {
		return "" // ?
	}

	links := make([]string, len(matches))

	for i := 0; i < len(links); i++ {
		m := matches[i]
		links[i] = twitterRegex.ReplaceAllString(m, "https://fxtwitter.com/$2/status/$4")
	}

	return strings.Join(links, "\n")
}

/*
func createMentions(m message) string {
	if mentions := m.content.Mentions; len(mentions) > 0 {
		out := make([]string, len(mentions))
		for i := 0; i < len(out); i++ {
			out[i] = mentions[i].Tag()
		}
		return "\nmentioned: " + strings.Join(out, ", ")
	}
	return ""
}*/

func hideEmbeds(s *session.Session, m message) {
	oldFlags := m.content.Flags
	newFlags := oldFlags | discord.SuppressEmbeds

	editMsgData := api.EditMessageData{
		Flags: &newFlags,
	}

	_, err := s.EditMessageComplex(m.content.ChannelID, m.content.ID, editMsgData)
	if err != nil {
		log.Println("error editing message:", err)
	}
}

func replaceMessage(s *session.Session, m message) {
	output := createOutput(m)
	//mentions := createMentions(m)
	hideEmbeds(s, m)

	newMessage := fmt.Sprintf(`%s` /*mentions,*/, output)

	_, err := s.SendMessageReply(m.content.ChannelID, newMessage, m.content.ID)
	//_, err := s.SendMessage(m.content.ChannelID, newMessage)
	if err != nil {
		log.Println("error sending message:", err)
	}

	/*if err = s.DeleteMessage(m.content.ChannelID, m.content.ID, api.AuditLogReason("fxtwitterbot")); err != nil {
		log.Println("error deleting message:", err)
	}*/
}

type message struct {
	content discord.Message
	author  discord.User
}

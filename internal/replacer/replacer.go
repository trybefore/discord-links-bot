package replacer

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

// Add adds a new replacer
func Add(replacer Replacer) {
	replacers = append(replacers, replacer)
}

const (
	Reddit        = "reddit"
	YoutubeShorts = "youtube_shorts"
	Instagram     = "instagram"
	Amazon        = "amazon"
	Discord       = "discord"
	Twitter       = "twitter"
	TikTok        = "tiktok"
	TikTokVM      = "tiktok_vm"
)

// ByName finds a replacer with the name, or nil
func ByName(name string) Replacer {
	filtered := Filter(replacers, func(e Replacer) bool {
		return e.Name() == name
	})

	if len(filtered) == 0 {
		return nil
	}

	return filtered[0]
}

var replacers []Replacer
var ErrNoMatch = errors.New("no regex match found")

type Replacer interface {
	Replace(string) (string, error)
	Matches(string) bool
	Name() string
}

func FindReplacers(m discord.Message) (out []Replacer) {
	return findReplacers(m)
}

func findReplacers(m discord.Message) (out []Replacer) {
	for _, replacer := range replacers {
		if replacer.Matches(m.Content) {
			out = append(out, replacer)
		}
	}

	return
}

// returns a new string with all regex matches replaced with replacement
func replaceMatches(regex *regexp.Regexp, message, replacement string) string {
	matches := regex.FindAllString(message, -1)
	if matches == nil {
		log.Printf("FindAllString is empty for string '%s' using (%s)", message, regex.String())
		return "" // ?
	}

	links := make([]string, len(matches))

	for i := 0; i < len(links); i++ {
		m := matches[i]
		link := regex.ReplaceAllString(m, replacement)
		if strings.Contains(message, "||") {
			link = fmt.Sprintf("||%s||", link)
		}

		links[i] = link
	}

	return strings.Join(links, "\n")
}

// ReplaceAll Goes through every replacer and replaces any matching strings, returning a new string with the matches replaced
func ReplaceAll(m Message) string {
	outputMessage := m.Content.Content

	for _, replacer := range *m.Replacers {
		newMessage, err := replacer.Replace(outputMessage)

		if !errors.Is(err, ErrNoMatch) && err != nil {
			log.Printf("error in replacer '%s': %v", replacer.Name(), err)
			continue
		}
		outputMessage = newMessage
	}

	return outputMessage
}

// hide embeds in message m
func hideEmbeds(s *state.State, m Message) {
	time.AfterFunc(time.Second*2, func() {
		oldFlags := m.Content.Flags
		newFlags := oldFlags | discord.SuppressEmbeds

		editMsgData := api.EditMessageData{
			Flags: &newFlags,
		}

		_, err := s.EditMessageComplex(m.Content.ChannelID, m.Content.ID, editMsgData)
		if err != nil {
			log.Println("error editing message:", err)
		}
	})
}

// SendReplacementMessage finds all matching strings and sends a new message in reply to Message m, with the matches replaced.
func SendReplacementMessage(s *state.State, m Message) {
	output := ReplaceAll(m)

	hideEmbeds(s, m)
	if _, err := s.SendMessageComplex(m.Content.ChannelID, api.SendMessageData{
		Content:         output,
		AllowedMentions: &api.AllowedMentions{},
		Reference: &discord.MessageReference{
			MessageID: m.Content.ID,
		},
	}); err != nil {
		log.Println("error sending reply message:", err)
	}
}

// Filter returns a new slice which includes any value returning true in fn from S
func Filter[S ~[]E, E any](s S, fn func(e E) bool) S {
	var out []E

	for _, e := range s {
		if fn(e) {
			out = append(out, e)
		}
	}

	return out
}

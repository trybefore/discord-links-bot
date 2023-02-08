package replacer

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/session"
)

var replacers []Replacer = []Replacer{
	amazonReplacer, twitterReplacer, discordReplacer, youtubeShortsReplacer, redditShortsReplacer,
}

var (
	amazonReplacer = &genericReplacer{
		regex:       regexp.MustCompile(`https?:\/\/(.*)\.amazon\.(de|com|co\.uk).*\/dp\/(\w*)`),
		replacement: "https://$1.amazon.$2/dp/$3",
	}
	twitterReplacer = &genericReplacer{
		regex:       regexp.MustCompile(`https?:\/\/(?P<tld>twitter)\.com\/(?:#!\/)?(\w+)\/status(es)?\/(\d+)`),
		replacement: "https://fxtwitter.com/$2/status/$4",
	}
	discordReplacer = &genericReplacer{
		regex:       regexp.MustCompile(`https?:\/\/media\.discordapp\.net/attachments/(\d+)/(\d+)/(.*)`),
		replacement: "https://cdn.discordapp.com/attachments/$1/$2/$3",
	}

	youtubeShortsReplacer = &genericReplacer{
		regex:       regexp.MustCompile(`https?:\/\/(?:www.)?youtube.com\/shorts\/(\w.*)`),
		replacement: "https://www.youtube.com/watch?v=$1",
	}

	redditShortsReplacer = &genericReplacer{
		regex:       regexp.MustCompile(`http(s)?(.+)reddit\.com/(?:r/)?([^/]+)/(?:(comments\/))?(\w{5,9})`),
		replacement: "https://www.reddit.com/$5",
	}
)

type Replacer interface {
	Replace(string) string
	Matches(string) bool
}

var _ Replacer = (*genericReplacer)(nil)

type genericReplacer struct {
	regex       *regexp.Regexp
	replacement string
}

func (t *genericReplacer) Replace(msg string) string {
	return replaceMatches(t.regex, msg, t.replacement)
}

func (t *genericReplacer) Matches(msg string) bool {
	return t.regex.MatchString(msg)
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

func ReplaceAll(m Message) string {
	outputMessage := m.Content.Content

	for _, replacer := range *m.Replacers {
		outputMessage = replacer.Replace(outputMessage)
	}

	return outputMessage
}
func hideEmbeds(s *session.Session, m Message) {
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
func ReplaceMessage(s *session.Session, m Message) {
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

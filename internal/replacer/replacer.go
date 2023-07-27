package replacer

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

var replacers []Replacer = []Replacer{
	amazon, twitter, dc, youtubeShorts, reddit,
}

var (
	amazon = &genericReplacer{
		regex:       regexp.MustCompile(`https?:\/\/(.*)\.amazon\.(de|com|co\.uk).*\/dp\/(\w*)`),
		replacement: "https://$1.amazon.$2/dp/$3",
	}
	twitter = &genericReplacer{
		regex:       regexp.MustCompile(`https?:\/\/(?P<tld>twitter)\.com\/(?:#!\/)?(\w+)\/status(es)?\/(\d+)`),
		replacement: "https://vxtwitter.com/$2/status/$4",
	}

	dc = &discordReplacer{
		regex: regexp.MustCompile(`https?:\/\/media\.discordapp\.net/attachments/(\d+)/(\d+)/(.*\.gif$)`),
		genericReplacer: &genericReplacer{
			regex:       regexp.MustCompile(`https?:\/\/media\.discordapp\.net/attachments/(\d+)/(\d+)/(.*$)`),
			replacement: "https://cdn.discordapp.com/attachments/$1/$2/$3",
		}}

	youtubeShorts = &genericReplacer{
		regex:       regexp.MustCompile(`https?:\/\/(?:www.)?youtube.com\/shorts\/(\w.*)`),
		replacement: "https://www.youtube.com/watch?v=$1",
	}

	reddit = &genericReplacer{
		regex:       regexp.MustCompile(`http(s)?://((old|www)\.)?reddit\.com/(?:r/)?(?P<subreddit>[^/]+)/(?:(comments\/))?(?P<submission>\w{5,9})(?P<comment>/.*/\w{3,9}/)?`),
		replacement: "https://www.reddit.com/r/${subreddit}/comments/${submission}${comment}",
	}
)

type Replacer interface {
	Replace(string) string
	Matches(string) bool
}

var _ Replacer = (*genericReplacer)(nil)

type discordReplacer struct {
	*genericReplacer
	regex *regexp.Regexp
}

func (t *discordReplacer) Replace(msg string) string {
	if t.regex.MatchString(msg) {
		return msg // link is a media.discordapp.com/.gif link, ignoring
	}

	if !t.genericReplacer.Matches(msg) {
		return msg
	}
	return replaceMatches(t.genericReplacer.regex, msg, t.genericReplacer.replacement)
}

type genericReplacer struct {
	regex       *regexp.Regexp
	replacement string
}

func (t *genericReplacer) Replace(msg string) string {
	if !t.Matches(msg) {
		return msg
	}
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
func ReplaceMessage(s *state.State, m Message) {
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

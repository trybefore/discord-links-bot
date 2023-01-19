package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/session"
)

var replacers []replacer

func init() {
	replacers = append(replacers, &genericReplacer{
		regex:       regexp.MustCompile(`https?:\/\/(?P<tld>twitter)\.com\/(?:#!\/)?(\w+)\/status(es)?\/(\d+)`),
		replacement: "https://fxtwitter.com/$2/status/$4",
	})
}

type replacer interface {
	Replace(string) string
	Matches(string) bool
}

var _ replacer = (*genericReplacer)(nil)

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

func findReplacers(m discord.Message) (out []replacer) {
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

func ReplaceAll(m message) string {
	outputMessage := m.content.Content

	for _, replacer := range *m.replacers {
		outputMessage = replacer.Replace(outputMessage)
	}

	return outputMessage
}

func replaceMessage(s *session.Session, m message) {
	output := ReplaceAll(m)

	hideEmbeds(s, m)

	_, err := s.SendMessageReply(m.content.ChannelID, output, m.content.ID)
	if err != nil {
		log.Println("error sending message:", err)
	}
}

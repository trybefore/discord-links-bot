package replacer

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"golang.org/x/sync/errgroup"
)

var replacers []Replacer = []Replacer{
	amazon, twitter, dc, youtubeShorts, reddit, instagram,
}

var (
	amazon = &genericReplacer{
		regex:       regexp.MustCompile(`https?://(.*)\.amazon\.(de|com|co\.uk).*/dp/(\w*)`),
		replacement: "https://$1.amazon.$2/dp/$3",
	}
	twitter = &genericReplacer{
		regex:       regexp.MustCompile(`https?://(?P<tld>twitter|x)\.com/(?:#!/)?(\w+)/status(es)?/(\d+)`),
		replacement: "https://vxtwitter.com/$2/status/$4",
	}

	dc = &discordReplacer{
		regex: regexp.MustCompile(`https?://media\.discordapp\.net/attachments/(\d+)/(\d+)/(.*\.gif$)`),
		genericReplacer: &genericReplacer{
			regex:       regexp.MustCompile(`https?://media\.discordapp\.net/attachments/(\d+)/(\d+)/(.*$)`),
			replacement: "https://cdn.discordapp.com/attachments/$1/$2/$3",
		}}

	youtubeShorts = &genericReplacer{
		regex:       regexp.MustCompile(`https?://(?:www.)?youtube.com/shorts/(\w.*)`),
		replacement: "https://www.youtube.com/watch?v=$1",
	}

	reddit = &linkFollower{
		regex:       regexp.MustCompile(`http(s)?://((old|www)\.)?reddit\.com/(?:r/)+(?P<subreddit>[^/]+)/(comments/|s/)?(?P<submission>\w{5,12})(?P<title>/\w+/)?(?P<comment>\w{3,9}(/)?)?`),
		replacement: "https://www.reddit.com/r/${subreddit}/comments/${submission}${comment}",
	}

	instagram = &genericReplacer{
		regex:       regexp.MustCompile(`http(s)://(\w{3}.)?instagram.com/reel(s)?/(?P<id>.*)/`),
		replacement: "https://www.ddinstagram.com/reel/${id}",
	}

	tiktok = &linkFollower{
		regex: regexp.MustCompile(`http(s)?://(())`),
	}
)

type Replacer interface {
	Replace(string) string
	Matches(string) bool
}

var _ Replacer = (*genericReplacer)(nil)
var _ Replacer = (*discordReplacer)(nil)
var _ Replacer = (*linkFollower)(nil)

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

type linkFollower struct {
	*genericReplacer
	regex       *regexp.Regexp
	replacement string
}

// Matches implements Replacer.
func (r *linkFollower) Matches(msg string) bool {
	return r.regex.MatchString(msg)
}

var linkFollowerClient = &http.Client{
	Timeout: time.Second * 5,
}

// Replace implements Replacer.
func (r *linkFollower) Replace(msg string) string {
	if !r.Matches(msg) {
		log.Printf("message doesn't match: %s", msg)
		return msg
	}

	matches := r.regex.FindAllString(msg, -1)
	if matches == nil {
		log.Printf("FindAllString is empty for string '%s' using (%s)", msg, r.regex.String())
		return "" // ?
	}

	links := make([]string, len(matches))

	for i := 0; i < len(links); i++ {
		link := matches[i]

		links[i] = link
	}

	followedLinks, err := followLinks(links)

	if err != nil {
		log.Printf("followLinks: %v", err)
		return msg
	}

	return strings.Join(followedLinks, "\n")
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
		outputMessage = replacer.Replace(outputMessage)
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

// followLinks Visits every link and returns the final (after redirect) destination for each one
func followLinks(links []string) (followedLinks []string, err error) {
	followChan := make(chan string, 1)

	grp, ctx := errgroup.WithContext(context.Background())
	grp.SetLimit(2)
	for _, link := range links {
		linkToFollow := link
		log.Printf("following link: %s", linkToFollow)
		grp.Go(func() error {
			req, err := http.NewRequest("GET", linkToFollow, nil)

			if err != nil {
				return err
			}

			req = req.WithContext(ctx)
			res, err := linkFollowerClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			followedURL := res.Request.URL
			followedURL.RawQuery = ""

			followedLink := followedURL.String()

			log.Printf("followed link %s -> %s", linkToFollow, followedLink)

			followChan <- followedLink

			return nil
		})
	}

	if err = grp.Wait(); err != nil {
		return
	}

	close(followChan)

	for followedLink := range followChan {
		followedLinks = append(followedLinks, followedLink)
	}

	return followedLinks, nil
}

package replacer

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var client = &http.Client{Timeout: time.Second * 15}

var _ Replacer = (*LinkFollower)(nil)

// LinkFollower follows matching links, and (optionally) does another regex+replace with the final url.
type LinkFollower struct {
	followRegex *regexp.Regexp // follow link if match

	destinationRegex       *regexp.Regexp // run on final url, unless nil
	destinationReplacement string

	name string
}

func NewLinkFollower(name string, followRegex *regexp.Regexp, destinationRegex *regexp.Regexp, destinationReplacement string) *LinkFollower {
	return &LinkFollower{followRegex: followRegex, destinationRegex: destinationRegex, destinationReplacement: destinationReplacement, name: name}
}

func (r *LinkFollower) Name() string {
	return r.name
}

func (r *LinkFollower) Replace(msg string) (string, error) {
	if !r.Matches(msg) {
		return msg, ErrNoMatch
	}

	matches := r.followRegex.FindAllString(msg, -1)

	if matches == nil {
		log.Printf("FindAllString is empty for string '%s' using (%s)", msg, r.followRegex.String())
		return "", ErrNoMatch // ?
	}

	links := make([]string, len(matches))

	for i := 0; i < len(links); i++ {
		link := matches[i]

		links[i] = link
	}

	followedLinks, err := followLinks(links)

	if err != nil {
		return msg, err
	}

	if r.destinationRegex == nil {
		return strings.Join(followedLinks, "\n"), nil
	}

	newLinks := make([]string, len(followedLinks))

	for i := 0; i < len(newLinks); i++ {
		link := followedLinks[i]

		newLinks[i] = replaceMatches(r.destinationRegex, link, r.destinationReplacement)
	}

	return strings.Join(newLinks, "\n"), nil
}

func (r *LinkFollower) Matches(msg string) bool {
	return r.followRegex.MatchString(msg)
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
				return fmt.Errorf("error creating request for link '%s': %w", linkToFollow, err)
			}

			req = req.WithContext(ctx)
			res, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("error following link '%s': %w", linkToFollow, err)
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

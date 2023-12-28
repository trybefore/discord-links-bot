package replacer

import (
	"regexp"
	"strings"
)

func init() {
	Add(
		NewDeferred(
			NewLinkFollower(
				Reddit,
				regexp.MustCompile(`http(s)?://((old|www)\.)?reddit\.com/(?:r/)+(?P<subreddit>[^/]+)/(comments/|s/)?(?P<submission>\w{5,12})(?P<title>/\w+/)?(?P<comment>\w{3,9}(/)?)?`),
				nil,
				"",
			),
			strings.NewReplacer("reddit.com/r/", "rxddit.com/r/", "www.reddit.com/r/", "rxddit.com/r/", "old.reddit.com/r/", "rxddit.com/r/"),
		),
	)
}

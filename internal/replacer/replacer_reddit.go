package replacer

import "regexp"

func init() {
	Add(
		NewLinkFollower(
			Reddit,
			regexp.MustCompile(`http(s)?://((old|www)\.)?reddit\.com/(?:r/)+(?P<subreddit>[^/]+)/(comments/|s/)?(?P<submission>\w{5,12})(?P<title>/\w+/)?(?P<comment>\w{3,9}(/)?)?`),
			nil,
			"",
		),
	)
}

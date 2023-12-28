package replacer

import "regexp"

func init() {
	Add(
		NewSimple(
			Instagram,
			regexp.MustCompile(`http(s)://(\w{3}.)?instagram.com/reel(s)?/(?P<id>.*)/`),
			"https://www.ddinstagram.com/reel/${id}",
		),
	)
}

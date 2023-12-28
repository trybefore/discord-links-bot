package replacer

import "regexp"

func init() {
	Add(
		NewSimple(
			Twitter,
			regexp.MustCompile(`https?://(?P<tld>twitter|x)\.com/(?:#!/)?(\w+)/status(es)?/(\d+)`),
			"https://vxtwitter.com/$2/status/$4",
		),
	)
}

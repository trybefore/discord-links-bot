package replacer

import "regexp"

func init() {
	Add(
		NewSimple(
			YoutubeShorts,
			regexp.MustCompile(`https?://(?:www.)?youtube.com/shorts/(\w.*)`),
			"https://www.youtube.com/watch?v=$1",
		),
	)
}

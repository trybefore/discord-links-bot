package replacer

import "regexp"

func init() {
	Add(
		NewSimple(
			Discord,
			regexp.MustCompile(`https?://media\.discordapp\.net/attachments/(\d+)/(\d+)/(.*[^.gif].$)`),
			"https://cdn.discordapp.com/attachments/$1/$2/$3",
		),
	)
}

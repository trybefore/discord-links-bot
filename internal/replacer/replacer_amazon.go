package replacer

import "regexp"

func init() {
	Add(
		NewSimple(
			Amazon,
			regexp.MustCompile(`https?://(.*)\.amazon\.(de|com|co\.uk).*/dp/(\w*)`),
			"https://$1.amazon.$2/dp/$3",
		),
	)
}

package replacer

import "regexp"

func init() {
	Add(
		NewLinkFollower(
			TikTok,
			regexp.MustCompile(`https://(?:www\.)?tiktok.com/(?P<username>@.+)/video/(?P<videoId>\d+)(?:\?|[\s]|$)?`),
			regexp.MustCompile(`https://(?:www\.)?tiktok.com/(?P<username>@.+)/video/(?P<videoId>\d+)(?:\?|[\s]|$)?`),
			"https://www.vxtiktok.com/${username}/video/${videoId}",
		),
	)

	Add(
		NewLinkFollower(
			TikTokVM,
			regexp.MustCompile(`https://vm.tiktok.com/(?P<videoId>\w+)(?:/)?(?:\?.*)?`),
			regexp.MustCompile(`https://(?:www\.)?tiktok.com/(?P<username>@.+)/video/(?P<videoId>\d+)(?:\?|[\s]|$)?`),
			"https://www.vxtiktok.com/${username}/video/${videoId}",
		),
	)
}

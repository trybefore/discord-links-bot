package main

import (
	"regexp"
	"testing"
)

func TestTwitterRegex(t *testing.T) {
	replacer := &genericReplacer{regex: regexp.MustCompile(`https?:\/\/(?P<tld>twitter)\.com\/(?:#!\/)?(\w+)\/status(es)?\/(\d+)`), replacement: "https://fxtwitter.com/$2/status/$4"}

	tests := []struct {
		replacer *genericReplacer
		want     string
		have     string
	}{
		{
			replacer: replacer,
			have:     "https://twitter.com/blablabla/status/12345678910",
			want:     "https://fxtwitter.com/blablabla/status/12345678910",
		},
		{
			replacer: replacer,
			have:     "https://twitter.com/blablabla/statuses/12345678910",
			want:     "https://fxtwitter.com/blablabla/status/12345678910",
		},
		{
			replacer: replacer,
			have:     "https://twitter.com/blablabla/statuses/12345678910?cheese=doodles",
			want:     "https://fxtwitter.com/blablabla/status/12345678910",
		},
		{
			replacer: replacer,
			have:     "http://twitter.com/blablabla/statuses/12345678910",
			want:     "https://fxtwitter.com/blablabla/status/12345678910",
		},
	}
	for _, tt := range tests {
		t.Run(tt.have, func(t *testing.T) {
			if got := tt.replacer.Replace(tt.have); got != tt.want {
				t.Fatalf("%s != %s", got, tt.want)
			}
		})
	}
}

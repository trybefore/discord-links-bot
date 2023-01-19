package main

import (
	"regexp"
	"testing"
)

func TestDiscordRegex(t *testing.T) {
	replacer := &genericReplacer{
		regex:       regexp.MustCompile(`https?:\/\/media\.discordapp\.net/attachments/(\d+)/(\d+)/(.*)`),
		replacement: "https://cdn.discordapp.com/attachments/$1/$2/$3",
	}
	tests := []struct {
		want string
		have string
	}{
		{
			have: "https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4",
			want: "https://cdn.discordapp.com/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4",
		},

		//https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4
		//https://cdn.discordapp.com/attachments/735399993485033472/1065376405921222686/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4
	}

	for _, tt := range tests {
		t.Run(tt.have, func(t *testing.T) {
			if got := replacer.Replace(tt.have); got != tt.want {
				t.Fatalf("%s != %s", got, tt.want)
			}
		})
	}
}

func TestTwitterRegex(t *testing.T) {
	replacer := &genericReplacer{regex: regexp.MustCompile(`https?:\/\/(?P<tld>twitter)\.com\/(?:#!\/)?(\w+)\/status(es)?\/(\d+)`), replacement: "https://fxtwitter.com/$2/status/$4"}

	tests := []struct {
		want string
		have string
	}{
		{
			have: "https://twitter.com/blablabla/status/12345678910",
			want: "https://fxtwitter.com/blablabla/status/12345678910",
		},
		{
			have: "https://twitter.com/blablabla/statuses/12345678910",
			want: "https://fxtwitter.com/blablabla/status/12345678910",
		},
		{
			have: "https://twitter.com/blablabla/statuses/12345678910?cheese=doodles",
			want: "https://fxtwitter.com/blablabla/status/12345678910",
		},
		{
			have: "http://twitter.com/blablabla/statuses/12345678910",
			want: "https://fxtwitter.com/blablabla/status/12345678910",
		},
	}
	for _, tt := range tests {
		t.Run(tt.have, func(t *testing.T) {
			if got := replacer.Replace(tt.have); got != tt.want {
				t.Fatalf("%s != %s", got, tt.want)
			}
		})
	}
}

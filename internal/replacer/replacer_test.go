package replacer

import (
	"fmt"
	"testing"
)

func TestRedditRegex(t *testing.T) {
	tests := []struct {
		want string
		have string
	}{
		{
			have: "https://www.reddit.com/r/WTF/comments/17j3sd/this_crawled_into_one_of_my_students_pool_and_she/c867wpi/#c867wpi",
			want: "https://www.reddit.com/17j3sd",
		},
		{
			have: "https://www.reddit.com/r/reddit.com/comments/17863/two_countries_one_booming_one_struggling_which/c13/",
			want: "https://www.reddit.com/17863",
		},
		{
			have: "https://old.reddit.com/r/switcharoo/comments/u24xnc/bond_girl_vs_james_bond/j3g066i/",
			want: "https://www.reddit.com/u24xnc",
		},
		{
			have: `https://www.reddit.com/r/truetf2/comments/107nizk/what_makes_a_lot_of_the_configuration/
			`,
			want: `https://www.reddit.com/107nizk`,
		},
		{
			have: `https://www.reddit.com/r/truetf2/comments/107nizk/what_makes_a_lot_of_the_configuration/
			https://www.reddit.com/r/truetf2/comments/asdf123/what_makes_a_lot_of_the_configuration/
			https://www.reddit.com/r/truetf2/comments/test123/what_makes_a_lot_of_the_configuration/
			`,
			want: `https://www.reddit.com/107nizk
https://www.reddit.com/asdf123
https://www.reddit.com/test123`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got := redditShortsReplacer.Replace(tt.have); got != tt.want {
				t.Fatalf("[%s] %s != %s", tt.have, got, tt.want)
			}
		})
	}
}

func TestYoutubeShortsRegex(t *testing.T) {
	tests := []struct {
		want string
		have string
	}{
		{
			have: "https://www.youtube.com/shorts/u3juys4lq-E",
			want: "https://www.youtube.com/watch?v=u3juys4lq-E",
		},
		{
			have: "https://youtube.com/shorts/u3juys4lq-E",
			want: "https://www.youtube.com/watch?v=u3juys4lq-E",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got := youtubeShortsReplacer.Replace(tt.have); got != tt.want {
				t.Fatalf("[%s] %s != %s", tt.have, got, tt.want)
			}
		})
	}
}

func TestAmazonRegex(t *testing.T) {
	tests := []struct {
		want string
		have string
	}{
		{
			have: "https://www.amazon.com/dp/B09B1HMJ9Z/ref=as_li_ss_tl?ie=UTF8&smid=ATVPDKIKX0DER&th=1&linkCode=sl1&tag=sec2002-20",
			want: "https://www.amazon.com/dp/B09B1HMJ9Z",
		},
		{
			have: "https://www.amazon.co.uk/dp/B09B1HMJ9Z/ref=as_li_ss_tl?ie=UTF8&smid=ATVPDKIKX0DER&th=1&linkCode=sl1&tag=sec2002-20",
			want: "https://www.amazon.co.uk/dp/B09B1HMJ9Z",
		},

		{
			have: "https://www.amazon.de/dp/B09B1HMJ9Z/ref=as_li_ss_tl?ie=UTF8&smid=ATVPDKIKX0DER&th=1&linkCode=sl1&tag=sec2002-20",
			want: "https://www.amazon.de/dp/B09B1HMJ9Z",
		},
		{
			have: "https://www.amazon.com/Monster-High-School-Playset/dp/B006O6F932/ref=gbph_tit_e-7_fb02_85d3d028?smid=A3CXJV2JYTL237&pf_rd_p=8e268714-ad3d-444b-b0df-d51d8825fb02&pf_rd_s=events-center-c-7&pf_rd_t=701&pf_rd_i=HTL_desktop&pf_rd_m=ATVPDKIKX0DER&pf_rd_r=8MKN8SY6C5ZP4NC1C0RB",
			want: "https://www.amazon.com/dp/B006O6F932",
		},
		{
			have: `https://www.amazon.com/Monster-High-School-Playset/dp/B006O6F932/ref=gbph_tit_e-7_fb02_85d3d028?smid=A3CXJV2JYTL237&pf_rd_p=8e268714-ad3d-444b-b0df-d51d8825fb02&pf_rd_s=events-center-c-7&pf_rd_t=701&pf_rd_i=HTL_desktop&pf_rd_m=ATVPDKIKX0DER&pf_rd_r=8MKN8SY6C5ZP4NC1C0RB
https://www.amazon.com/Monster-High-School-Playset/dp/B006O6F932/ref=gbph_tit_e-7_fb02_85d3d028?smid=A3CXJV2JYTL237&pf_rd_p=8e268714-ad3d-444b-b0df-d51d8825fb02&pf_rd_s=events-center-c-7&pf_rd_t=701&pf_rd_i=HTL_desktop&pf_rd_m=ATVPDKIKX0DER&pf_rd_r=8MKN8SY6C5ZP4NC1C0RB
https://www.amazon.com/Monster-High-School-Playset/dp/B006O6F932/ref=gbph_tit_e-7_fb02_85d3d028?smid=A3CXJV2JYTL237&pf_rd_p=8e268714-ad3d-444b-b0df-d51d8825fb02&pf_rd_s=events-center-c-7&pf_rd_t=701&pf_rd_i=HTL_desktop&pf_rd_m=ATVPDKIKX0DER&pf_rd_r=8MKN8SY6C5ZP4NC1C0RB`,
			want: `https://www.amazon.com/dp/B006O6F932
https://www.amazon.com/dp/B006O6F932
https://www.amazon.com/dp/B006O6F932`,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got := amazonReplacer.Replace(tt.have); got != tt.want {
				t.Fatalf("%s != %s", got, tt.want)
			}
		})
	}
}

func TestDiscordRegex(t *testing.T) {
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

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got := discordReplacer.Replace(tt.have); got != tt.want {
				t.Fatalf("%s != %s", got, tt.want)
			}
		})
	}
}

func TestTwitterRegex(t *testing.T) {
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
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got := twitterReplacer.Replace(tt.have); got != tt.want {
				t.Fatalf("%s != %s", got, tt.want)
			}
		})
	}
}

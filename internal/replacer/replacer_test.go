package replacer

import (
	"errors"
	"fmt"
	"testing"
)

func TestTikTokRegex(t *testing.T) {
	t.Parallel()
	ttReplacer := ByName(TikTok)
	vmReplacer := ByName(TikTokVM)

	tests := []struct {
		want     string
		have     string
		replacer Replacer
	}{
		{
			have:     "https://vm.tiktok.com/ZM6BdBuuY/",
			want:     "https://www.vxtiktok.com/@crowndefend/video/7311691285578943776",
			replacer: vmReplacer,
		},
		{
			have:     "https://www.tiktok.com/@realcompmemer/video/7314546788617309471",
			want:     "https://www.vxtiktok.com/@realcompmemer/video/7314546788617309471",
			replacer: ttReplacer,
		},
		{
			have:     "https://www.tiktok.com/@butt.erhand/video/7310948082781375745",
			want:     "https://www.vxtiktok.com/@butt.erhand/video/7310948082781375745",
			replacer: ttReplacer,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got, err := tt.replacer.Replace(tt.have); got != tt.want && err == nil {
				t.Fatalf("%s != %s", got, tt.want)
			} else if err != nil && !errors.Is(err, ErrNoMatch) {
				t.Fatalf("error processing link '%s': %v", tt.have, err)
			}
		})
	}
}

func TestRedditRegex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		want string
		have string
	}{
		{
			have: "https://www.reddit.com/r/WTF/comments/17j3sd/this_crawled_into_one_of_my_students_pool_and_she/c867wpi/#c867wpi",
			want: "https://www.reddit.com/r/WTF/comments/17j3sd/this_crawled_into_one_of_my_students_pool_and_she/c867wpi/",
		},
		{
			have: "https://www.reddit.com/r/reddit.com/comments/17863/two_countries_one_booming_one_struggling_which/c13",
			want: "https://www.reddit.com/r/reddit.com/comments/17863/two_countries_one_booming_one_struggling_which/c13/",
		},
		{
			have: "https://old.reddit.com/r/switcharoo/comments/u24xnc/bond_girl_vs_james_bond/j3g066i/",
			want: "https://old.reddit.com/r/switcharoo/comments/u24xnc/bond_girl_vs_james_bond/j3g066i/",
		},
		{
			have: `https://www.reddit.com/r/truetf2/comments/107nizk/what_makes_a_lot_of_the_configuration?baba123=dafsdfasdf`,
			want: `https://www.reddit.com/r/truetf2/comments/107nizk/`,
		},
		{
			have: "https://reddit.com/r/pchelp/s/GPxofho2iY",
			want: "https://www.reddit.com/r/pchelp/comments/167kev4/i_kinda_messed_up_should_i_redo_it/",
		},
		{
			have: "https://reddit.com/r/computers/s/BN4uCFXFyC",
			want: "https://www.reddit.com/r/computers/comments/16o6g5r/how_much_dose_my_pc_worth/",
		},
	}

	replacer := ByName(Reddit)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got, err := replacer.Replace(tt.have); got != tt.want && err == nil {
				t.Fatalf("%s != %s", got, tt.want)
			} else if err != nil && !errors.Is(err, ErrNoMatch) {
				t.Fatalf("error processing link '%s': %v", tt.have, err)
			}
		})
	}
}

func TestYoutubeShortsRegex(t *testing.T) {
	t.Parallel()
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

	replacer := ByName(YoutubeShorts)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got, err := replacer.Replace(tt.have); got != tt.want && err == nil {
				t.Fatalf("%s != %s", got, tt.want)
			} else if err != nil && !errors.Is(err, ErrNoMatch) {
				t.Fatalf("error processing link '%s': %v", tt.have, err)
			}
		})
	}
}

func TestInstagramRegex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		want string
		have string
	}{
		{
			have: "https://www.instagram.com/reel/CztdYC8ryw7/?igshid=abcdefghujkl==",
			want: "https://www.ddinstagram.com/reel/CztdYC8ryw7",
		},
		{
			have: "https://www.instagram.com/reel/CzmhWrGNL9u/?utm_source=ig_web_copy_link",
			want: "https://www.ddinstagram.com/reel/CzmhWrGNL9u",
		},
		{
			have: "https://www.instagram.com/reels/CzmhWrGNL9u/",
			want: "https://www.ddinstagram.com/reel/CzmhWrGNL9u",
		},
	}

	replacer := ByName(Instagram)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got, err := replacer.Replace(tt.have); got != tt.want && err == nil {
				t.Fatalf("%s != %s", got, tt.want)
			} else if err != nil && !errors.Is(err, ErrNoMatch) {
				t.Fatalf("error processing link '%s': %v", tt.have, err)
			}
		})
	}
}

func TestAmazonRegex(t *testing.T) {
	t.Parallel()
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

	replacer := ByName(Amazon)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got, err := replacer.Replace(tt.have); got != tt.want && err == nil {
				t.Fatalf("%s != %s", got, tt.want)
			} else if err != nil && !errors.Is(err, ErrNoMatch) {
				t.Fatalf("error processing link '%s': %v", tt.have, err)
			}
		})
	}
}

func TestDiscordRegex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		want string
		have string
	}{
		{
			have: "https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4",
			want: "https://cdn.discordapp.com/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4",
		},
		{
			have: "https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.gif",
			want: "https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.gif",
		},
		{
			have: "https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.png",
			want: "https://cdn.discordapp.com/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.png",
		},
		{
			have: "https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.jpeg",
			want: "https://cdn.discordapp.com/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.jpeg",
		},
		{
			have: "https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.jpg",
			want: "https://cdn.discordapp.com/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.jpg",
		},
		//https://media.discordapp.net/attachments/483348725704556557/1065345579762335915/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4
		//https://cdn.discordapp.com/attachments/735399993485033472/1065376405921222686/v12044gd0000cf3g5rrc77u1ikgnhp8g.mp4
	}

	replacer := ByName(Discord)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got, err := replacer.Replace(tt.have); got != tt.want && err == nil {
				t.Fatalf("%s != %s", got, tt.want)
			} else if err != nil && !errors.Is(err, ErrNoMatch) {
				t.Fatalf("error processing link '%s': %v", tt.have, err)
			}
		})
	}
}

func TestTwitterRegex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		want string
		have string
	}{
		{
			have: "https://twitter.com/blablabla/status/12345678910",
			want: "https://vxtwitter.com/blablabla/status/12345678910",
		},
		{
			have: "https://twitter.com/blablabla/statuses/12345678910",
			want: "https://vxtwitter.com/blablabla/status/12345678910",
		},
		{
			have: "https://twitter.com/blablabla/statuses/12345678910?cheese=doodles",
			want: "https://vxtwitter.com/blablabla/status/12345678910",
		},
		{
			have: "http://twitter.com/blablabla/statuses/12345678910",
			want: "https://vxtwitter.com/blablabla/status/12345678910",
		},
	}
	replacer := ByName(Twitter)

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			if got, err := replacer.Replace(tt.have); got != tt.want && err == nil {
				t.Fatalf("%s != %s", got, tt.want)
			} else if err != nil && !errors.Is(err, ErrNoMatch) {
				t.Fatalf("error processing link '%s': %v", tt.have, err)
			}
		})
	}
}

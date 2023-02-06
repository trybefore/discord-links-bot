package replacer

import (
	"github.com/diamondburned/arikawa/v3/discord"
)

type Message struct {
	Content   discord.Message
	Author    discord.User
	Replacers *[]Replacer
}

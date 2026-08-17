package newsletter

import (
	th "github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego"
)

type Newsletter struct {
	bot *telego.Bot
}

func New(bot *telego.Bot) *Newsletter {
	return &Newsletter{bot}
}

func (n *Newsletter) BroadcastMessage(
	ctx *th.Context,
	recipients []telego.ChatID,
	message string,
) {
	for _, r := range recipients {
		n.bot.SendMessage(ctx, &telego.SendMessageParams{
			ChatID: r,
			Text: message,
		})	
	}
}

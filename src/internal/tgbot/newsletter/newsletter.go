package newsletter

import (
	"context"

	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/mymmrac/telego"
)

type Newsletter struct {
	bot *telego.Bot
}

func New(bot *telego.Bot) *Newsletter {
	return &Newsletter{bot}
}

func (n *Newsletter) SendPaymentNotification(
	ctx context.Context,
	recipient int64,
	message string,
) {
	confirmPaymentKb := tu.InlineKeyboard([][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Подтвердить оплату").
				WithCallbackData("confirm_payment"),
		),
	}...)

	n.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: recipient},
		Text: message,
		ReplyMarkup: confirmPaymentKb,
	})	
}

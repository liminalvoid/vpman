package newsletter

import (
	"context"
	"log/slog"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

type Newsletter struct {
	log *slog.Logger
	bot *telego.Bot
}

func New(log *slog.Logger, bot *telego.Bot) *Newsletter {
	return &Newsletter{log, bot}
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

	n.log.DebugContext(
		ctx,
		"sent payment notification",
		"message",
		message,
	)

	n.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID:      telego.ChatID{ID: recipient},
		Text:        message,
		ReplyMarkup: confirmPaymentKb,
	})
}

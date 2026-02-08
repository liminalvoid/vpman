package handlers

import (
	kb "github.com/liminalvoid/vpn-manager/internal/keyboards"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func (h *Handlers) HandleStart(ctx *th.Context, update telego.Update) error {
	bot := h.bot

	bot.SendMessage(
		ctx,
		tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Привет, %s\n\nКоличество ключей: 5/5\n\nДоступные локации:\n\n🇩🇪 Германия\n🇫🇷 Франция\n🇳🇱 Нидерланды",
			update.Message.From.FirstName,
		).
			WithReplyMarkup(kb.Start),
	)

	return nil
}

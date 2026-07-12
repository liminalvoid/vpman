package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

var startKb = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Сгенерировать ключ").
			WithCallbackData("wizard:generate_key"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Управление ключами").
			WithCallbackData("wizard:manage_keys"),
	),
)

func (h *Handlers) HandleStart(ctx *th.Context, update telego.Update) error {
	h.bot.SendMessage(
		ctx,
		tu.Messagef(
			tu.ID(update.Message.Chat.ID),
			"Привет, %s\n\nКоличество ключей: 5/5\n\nДоступные локации:\n\n🇩🇪 Германия\n🇫🇷 Франция\n🇳🇱 Нидерланды",
			update.Message.From.FirstName,
		).
			WithReplyMarkup(startKb),
	)

	return nil
}

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
	var chatID telego.ChatID
	var firstName string

	switch {
	case update.CallbackQuery != nil:
		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		chatID = update.CallbackQuery.Message.GetChat().ChatID()
		firstName = update.CallbackQuery.From.FirstName
	case update.Message != nil:
		chatID = update.Message.Chat.ChatID()
		firstName = update.Message.From.FirstName
	}

	h.bot.SendMessage(
		ctx,
		tu.Messagef(
			chatID,
			"Привет, %s\n\nКоличество ключей: 5/5\n\nДоступные локации:\n\n🇩🇪 Германия\n🇫🇷 Франция\n🇳🇱 Нидерланды",
			firstName,
		).
			WithReplyMarkup(startKb),
	)

	return nil
}

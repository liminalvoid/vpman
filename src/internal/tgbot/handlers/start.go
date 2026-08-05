package handlers

import (
	"fmt"
	"strings"

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

	availableServers, err := h.serverService.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("querying servers: %w", err)
	}

	var formattedServers []string
	for _, s := range availableServers {
		formattedString := fmt.Sprintf(
			"%s %s | %s",
			convertCountryCodeToFlagEmoji(s.Location),
			s.LocationCountry,
			s.LocationCity,
		)
		formattedServers = append(formattedServers, formattedString)
	}

	configs, err := h.configService.ListConfigs(ctx, chatID.ID)
	if err != nil {
		return fmt.Errorf("querying configs: %w", err)
	}

	h.bot.SendMessage(
		ctx,
		tu.Messagef(
			chatID,
			"Привет, %s\n\nКоличество ключей: %d\n\nДоступные локации:\n\n%s",
			firstName,
			len(configs),
			strings.Join(formattedServers, "\n"),
		).
			WithReplyMarkup(startKb),
	)

	return nil
}

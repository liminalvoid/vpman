package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func renderStartKb(isAdmin bool) *telego.InlineKeyboardMarkup {
	baseKb := [][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Сгенерировать ключ").
				WithCallbackData("generate_key"),
		),
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Управление ключами").
				WithCallbackData("manage_keys"),
		),
	}
	adminKb := [][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Сгенерировать инвайт-код").
				WithCallbackData("admin:generate_invite"),
		),
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Установить дедлайн по оплате").
				WithCallbackData("admin:set_past_due"),
		),
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Управление пользователями").
				WithCallbackData(ManageUsersPrefix),
		),
	}

	if isAdmin {
		baseKb = append(baseKb, adminKb...)
	}

	return tu.InlineKeyboard(baseKb...)
}

func (h *Handlers) HandleStart(ctx *th.Context, update telego.Update) error {
	var chatID telego.ChatID
	var userID int64
	var firstName string

	switch {
	case update.CallbackQuery != nil:
		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		chatID = update.CallbackQuery.Message.GetChat().ChatID()
		firstName = update.CallbackQuery.From.FirstName
		userID = update.CallbackQuery.From.ID
	case update.Message != nil:
		chatID = update.Message.Chat.ChatID()
		firstName = update.Message.From.FirstName
		userID = update.Message.From.ID
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

	subscription, err := h.subscriptionService.GetByTelegramID(ctx, userID)
	if err != nil {
		return fmt.Errorf("retrieving subscriptions: %w", err)
	}
	maxConfigCount := subscription.Quantity

	isAdmin := h.sessions.Get(userID).IsAdmin
	h.bot.SendMessage(
		ctx,
		tu.Messagef(
			chatID,
			"Привет, %s\n\nКоличество ключей: %d/%d\n\nДоступные локации:\n\n%s",
			firstName,
			len(configs),
			maxConfigCount,
			strings.Join(formattedServers, "\n"),
		).
			WithReplyMarkup(renderStartKb(isAdmin)),
	)

	return nil
}

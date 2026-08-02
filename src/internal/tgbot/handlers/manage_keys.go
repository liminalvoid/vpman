package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

var manageKeyKb = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Изменить название").
			WithCallbackData("manage_keys:edit_name"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Удалить").
			WithCallbackData("manage_keys:delete"),
	),
)

func (h *Handlers) HandleManageKeys(ctx *th.Context, update telego.Update) error {
	var userId int64
	switch {
	case update.CallbackQuery != nil:
		userId = update.CallbackQuery.From.ID
	case update.Message != nil:
		userId = update.Message.From.ID
	default:
		return nil
	}

	response := Response{}

	h.mutex.RLock()
	key := h.keys[userId]
	h.mutex.RUnlock()

	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		switch {
		case strings.Contains(update.CallbackQuery.Data, "key"):
			keyId := strings.ReplaceAll(update.CallbackQuery.Data, "manage_keys:key-", "")
			key.KeyID = keyId

			response.Text = fmt.Sprintf("**%s**\n\nЛокация: %s\nПротокол: %s\nСтатус: активен", key.Name, "EU", key.Protocol)
			response.ReplyMarkup = manageKeyKb
		case strings.Contains(update.CallbackQuery.Data, "edit"):
			response.Text = "Введите новое название"
		case strings.Contains(update.CallbackQuery.Data, "delete"):
			response.Text = "Конфиг удален"
		}
	case update.Message != nil:
		response.ChatID = update.Message.Chat.ChatID()
	}

	configs, err := h.peerService.ListConfigs(ctx, response.ChatID.ID)
	if err != nil {
		return fmt.Errorf("querying configs: %w", err)
	}

	keysKbButtons := make([][]telego.InlineKeyboardButton, 0)

	for _, c := range configs {
		keyId := strings.Replace(strings.ToLower(c.Name), " ", "-", 1)

		buttonText := fmt.Sprintf(
			"%s, %s",
			c.Name,
			c.Protocol,
		)

		keysKbButtons = append(
			keysKbButtons,
			tu.InlineKeyboardRow(
				tu.
					InlineKeyboardButton(buttonText).
					WithCallbackData(fmt.Sprintf("manage_keys:key-%s", keyId)),
			),
		)
	}

	keysKb := tu.InlineKeyboard(keysKbButtons...)

	if update.CallbackQuery != nil {
		switch update.CallbackQuery.Data {
		case "wizard:manage_keys":
			response.Text = "Текущие ключи"
			response.ReplyMarkup = keysKb
		}
	}

	if response.Text == "" {
		return nil
	}

	message := tu.Message(response.ChatID, response.Text)

	if response.ReplyMarkup != nil {
		message = message.WithReplyMarkup(response.ReplyMarkup)
	}

	if _, err := h.bot.SendMessage(ctx, message); err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	h.mutex.RLock()
	h.keys[userId] = key
	h.mutex.RUnlock()

	return nil
}

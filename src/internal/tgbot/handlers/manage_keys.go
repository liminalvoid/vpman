package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type Config struct {
	Name     string
	Location string
	Protocol string
	Key      string
}

func (h *Handlers) HandleManageKeys(ctx *th.Context, update telego.Update) error {
	response := Response{}

	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)
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

	return nil
}

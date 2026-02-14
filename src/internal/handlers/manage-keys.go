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
	userId := int64(0)
	switch {
	case update.CallbackQuery != nil:
		userId = update.CallbackQuery.From.ID
	case update.Message != nil:
		userId = update.Message.From.ID
	}

	// FIXME: Think if this `if` is necessary
	if userId == 0 {
		return nil
	}

	currentKeys := [...]*Config{
		&Config{
			Name: "VPN",
			Location: "fr",
			Protocol: "AWG",
			Key: "alkdsjflkwo3283",
		},
		&Config{
			Name: "VPN 2",
			Location: "nl",
			Protocol: "AWG",
			Key: "alkdsjflkwo3283",
		},
		&Config{
			Name: "VPN 3",
			Location: "ge",
			Protocol: "XRay",
			Key: "alkdsjflkwo3283",
		},
	}

	keysKbButtons := make([][]telego.InlineKeyboardButton, 0)

	for i := range len(currentKeys) {
		keyId := strings.ToLower(currentKeys[i].Name)
		keyId = strings.ReplaceAll(keyId, " ", "-")

		buttonText := fmt.Sprintf(
			"%s, %s, %s",
			currentKeys[i].Name,
			locationMap[currentKeys[i].Location],
			currentKeys[i].Protocol,
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

	response := Response{}

	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		switch update.CallbackQuery.Data {
		case "wizard:manage_keys":
			response.Text = "Текущие ключи"
			response.ReplyMarkup = keysKb
		}
	}

	message := tu.Message(response.ChatID, response.Text)

	if response.ReplyMarkup != nil {
		message = message.WithReplyMarkup(response.ReplyMarkup)
	}

	h.bot.SendMessage(
		ctx,
		message,
	)

	return nil
}

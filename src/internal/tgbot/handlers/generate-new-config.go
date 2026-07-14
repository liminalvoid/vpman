package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type Response struct {
	Text        string
	ReplyMarkup telego.ReplyMarkup
	ChatID      telego.ChatID
}

type GenerateConfigState int

const (
	StateCreateConfig GenerateConfigState = iota
	StateCreateConfigEnterKeyName
	StateCreateConfigSelectLocation
	StateCreateConfigSelectProtocol
	StateCreateConfigConfirm
)

var locationMap = map[string]string{
	"ge": "🇩🇪 Германия",
	"fr": "🇫🇷 Франция",
	"nl": "🇳🇱 Нидерланды",
}

var protocolMap = map[string]string{
	"awg":  "Amnezia WG",
	"xray": "XRay",
}

var submitActionKb = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Да").
			WithCallbackData("generate_key:yes"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Изменить").
			WithCallbackData("generate_key:edit"),
	),
)

var selectLocationKb = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton(locationMap["ge"]).
			WithCallbackData("generate_key:location-ge"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton(locationMap["fr"]).
			WithCallbackData("generate_key:location-fr"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton(locationMap["nl"]).
			WithCallbackData("generate_key:location-nl"),
	),
)

var selectProtocolKb = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton(protocolMap["awg"]).
			WithCallbackData("generate_key:protocol-awg"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton(protocolMap["xray"]).
			WithCallbackData("generate_key:protocol-xray"),
	),
)

var finalMenuKb = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("В меню").
			WithCallbackData("menu"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Сгенерировать ещё раз").
			WithCallbackData("wizard:generate_key"),
	),
)

func (h *Handlers) HandleNewConfig(ctx *th.Context, update telego.Update) error {
	userId := int64(0)
	switch {
	case update.CallbackQuery != nil:
		userId = update.CallbackQuery.From.ID
	case update.Message != nil:
		userId = update.Message.From.ID
	default:
		return nil
	}

	h.mutex.RLock()
	key := h.keys[userId]
	h.mutex.RUnlock()

	response := Response{}

	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		switch {
		case update.CallbackQuery.Data == "wizard:generate_key" || update.CallbackQuery.Data == "generate_key:edit":
			response.Text = "Введи название ключа"
			key.State = StateCreateConfigEnterKeyName
		case strings.Contains(update.CallbackQuery.Data, "location"):
			key.Location = strings.ReplaceAll(update.CallbackQuery.Data, "generate_key:location-", "")
			key.State = StateCreateConfigConfirm

			location := locationMap[key.Location]

			response.Text = fmt.Sprintf(
				"Название: %s\nПротокол: %s\nЛокация: %s\n\nСоздать ключ?",
				key.Name,
				key.Protocol,
				location,
			)
			response.ReplyMarkup = submitActionKb
		case strings.Contains(update.CallbackQuery.Data, "protocol"):
			key.Protocol = strings.ReplaceAll(update.CallbackQuery.Data, "generate_key:protocol-", "")
			key.Protocol = protocolMap[key.Protocol]
			key.State = StateCreateConfigSelectProtocol

			response.Text = "Выбери локацию"
			response.ReplyMarkup = selectLocationKb
		case update.CallbackQuery.Data == "generate_key:yes":
			config, err := h.ps.CreatePeer(ctx, update.CallbackQuery.From.ID, key.Name, key.Location)
			if err != nil {
				response.Text = fmt.Sprintf("Не удалось создать конфиг: %s", err)
				response.ReplyMarkup = finalMenuKb
				key.State = StateCreateConfig

				break
			}

			response.Text = fmt.Sprintf("Ключ успешно создан\\!\n\n```plain\n%s\n```", config)
			response.ReplyMarkup = finalMenuKb
			key.State = StateCreateConfig
		}
	case update.Message != nil:
		response.ChatID = update.Message.Chat.ChatID()

		switch key.State {
		case StateCreateConfigEnterKeyName:
			key.Name = update.Message.Text
			key.State = StateCreateConfigSelectProtocol
			response.Text = "Выбери протокол"
			response.ReplyMarkup = selectProtocolKb
		}
	}

	message := tu.Message(response.ChatID, response.Text).WithParseMode(telego.ModeMarkdownV2)

	if response.ReplyMarkup != nil {
		message = message.WithReplyMarkup(response.ReplyMarkup)
	}

	h.bot.SendMessage(
		ctx,
		message,
	)

	h.mutex.Lock()
	h.keys[userId] = key
	h.mutex.Unlock()

	return nil
}

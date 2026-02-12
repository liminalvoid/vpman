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

type State int

const (
	StateDefault State = iota
	StateEnterKeyName
	StateSelectLocation
	StateSelectProtocol
	StateConfirm
)

var locationMap = map[string]string{
	"ge": "🇩🇪 Германия",
	"fr": "🇫🇷 Франция",
	"nl": "🇳🇱 Нидерланды",
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
			InlineKeyboardButton("🇩🇪 Германия").
			WithCallbackData("generate_key:location-ge"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("🇫🇷 Франция").
			WithCallbackData("generate_key:location-fr"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("🇳🇱 Нидерланды").
			WithCallbackData("generate_key:location-nl"),
	),
)

func (h *Handlers) HandleNewConfig(ctx *th.Context, update telego.Update) error {
	userId := int64(0)
	switch {
	case update.CallbackQuery != nil:
		userId = update.CallbackQuery.From.ID
	case update.Message != nil:
		userId = update.Message.From.ID
	}

	if userId == 0 {
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
		case update.CallbackQuery.Data == "wizard:generate_key":
			response.Text = "Введи название ключа"
			key.State = StateEnterKeyName
		case update.CallbackQuery.Data == "generate_key:yes":
			response.Text = "Ключ успешно создан"
			key.State = StateDefault
		case strings.Contains(update.CallbackQuery.Data, "location"):
			key.Location = strings.Replace(update.CallbackQuery.Data, "generate_key:location-", "", 1)
			key.State = StateConfirm

			location := locationMap[key.Location]

			response.Text = fmt.Sprintf(
				"Название: %s\nПротокол: %s\nЛокация: %s\n\nСоздать ключ?",
				key.Name,
				key.Protocol,
				location,
			)
			response.ReplyMarkup = submitActionKb
		}
	case update.Message != nil:
		response.ChatID = update.Message.Chat.ChatID()

		switch key.State {
		case StateEnterKeyName:
			key.Name = update.Message.Text
			key.State = StateSelectProtocol
			response.Text = "Выбери протокол"
		case StateSelectProtocol:
			key.Protocol = update.Message.Text
			key.State = StateSelectLocation
			response.Text = "Выбери локацию"
			response.ReplyMarkup = selectLocationKb
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

	h.mutex.Lock()
	h.keys[userId] = key
	h.mutex.Unlock()

	return nil
}

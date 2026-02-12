package handlers

import (
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type Response struct {
	Text        string
	ReplyMarkup telego.ReplyMarkup
}

type State int

const (
	StateDefault State = iota
	StateEnterKeyName
	StateSelectLocation
	StateSelectProtocol
	StateConfirm
)

var onSubmitActionKb = tu.InlineKeyboard(
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

	switch {
	case update.CallbackQuery != nil:
		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		text := ""
		h.bot.Logger().Debugf("ballllllllllllllls")

		switch update.CallbackQuery.Data {
		case "wizard:generate_key":
			text = "Введи название ключа"
			key.State = StateEnterKeyName
		case "generate_key:yes":
			text = "Ключ успешно создан"
			key.State = StateDefault
		}

		h.bot.SendMessage(
			ctx,
			tu.Message(
				update.CallbackQuery.Message.GetChat().ChatID(),
				text,
			),
		)
	case update.Message != nil:
		response := Response{}

		switch key.State {
		case StateEnterKeyName:
			key.Name = update.Message.Text
			key.State = StateSelectProtocol
			response.Text = "Выбери протокол"
		case StateSelectProtocol:
			key.Protocol = update.Message.Text
			key.State = StateSelectLocation
			response.Text = "Выбери локацию"
		case StateSelectLocation:
			key.Location = update.Message.Text
			key.State = StateConfirm
			response.Text = fmt.Sprintf(
				"Название: %s\nПротокол: %s\nЛокация: %s\n\nСоздать ключ?",
				key.Name,
				key.Protocol,
				key.Location,
			)
			response.ReplyMarkup = onSubmitActionKb
		}

		message := tu.Message(update.Message.Chat.ChatID(), response.Text)

		if response.ReplyMarkup != nil {
			message = message.WithReplyMarkup(response.ReplyMarkup)
		}

		h.bot.Logger().Debugf("%+v", message)

		h.bot.SendMessage(
			ctx,
			message,
		)
	}

	h.mutex.Lock()
	h.keys[userId] = key
	h.mutex.Unlock()

	return nil
}

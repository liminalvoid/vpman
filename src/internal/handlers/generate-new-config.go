package handlers

import (
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type State int

const (
	StateDefault State = iota
	StateEnterKeyName
	StateSelectProtocol
	StateConfirm
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

		text := "Введи название ключа"
		key.State = StateEnterKeyName

		h.bot.SendMessage(
			ctx,
			tu.Message(
				update.
					CallbackQuery.
					Message.
					GetChat().
					ChatID(),
				text,
			),
		)
	case update.Message != nil:
		text := ""

		switch key.State {
		case StateDefault:
		case StateEnterKeyName:
			key.Name = update.Message.Text
			key.State = StateSelectProtocol
			text = "Выбери протокол"
		case StateSelectProtocol:
			key.Protocol = update.Message.Text
			key.State = StateConfirm
			text = "Создать ключ?"
		case StateConfirm:
			if update.Message.Text == "Да" {
				text = "Ключ успешно создан"
			} else {
				key.State = StateDefault
			}
		}

		h.bot.SendMessage(ctx, tu.Message(update.Message.Chat.ChatID(), text))
	}

	h.mutex.Lock()
	h.keys[userId] = key
	h.mutex.Unlock()

	return nil
}

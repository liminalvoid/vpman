package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/tgbot/session"
)

var SetPastDuePrefix = fmt.Sprintf("%s:set_past_due", AdminHandlerPrefix)

var setPastDueKb = [][]telego.InlineKeyboardButton{
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Да").
			WithCallbackData(fmt.Sprintf("%s:yes", SetPastDuePrefix)),
	),
}

func (h *Handlers) HandleSetPastDue(
	ctx *th.Context,
	update telego.Update,
) error {
	var response Response
	var userID int64
	switch {
	case update.CallbackQuery != nil:
		userID = update.CallbackQuery.From.ID
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		callbackData := strings.ReplaceAll(
			update.CallbackQuery.Data,
			fmt.Sprintf("%s", SetPastDuePrefix),
			"",
		)
		switch callbackData {
		case "":
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateSetPastDue
			})
		case ":yes":
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateSetPastDueConfirm
			})
		}
	case update.Message != nil:
		userID = update.Message.From.ID
		response.ChatID = update.Message.Chat.ChatID()
	}

	var draftKb [][]telego.InlineKeyboardButton

	sess := h.sessions.Get(userID)
	switch sess.State{
	case session.StateSetPastDue:
		response.Text = "Установить дедлайн по оплате?"
		draftKb = setPastDueKb
	case session.StateSetPastDueConfirm:
		response.Text = "Успешно"
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateIdle
		})
	}

	if response.Text == "" {
		return nil
	}

	message := tu.Message(response.ChatID, response.Text).WithParseMode(telego.ModeMarkdownV2)
	response.ReplyMarkup = buildKbWithExitButton(draftKb...)
	message = message.WithReplyMarkup(response.ReplyMarkup)

	h.bot.SendMessage(
		ctx,
		message,
	)

	return nil
}

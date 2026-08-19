package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/tgbot/session"
)

var ManageUsersPrefix = fmt.Sprintf("%s:manage_users", AdminHandlerPrefix)

func (h *Handlers) HandleManageUsers(
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
				s.State = session.StateManageUsers
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
	switch sess.State {
	case session.StateManageUsers:
		users, err := h.userService.ListAll(ctx)
		if err != nil {
			return fmt.Errorf("retrieving users: %w", err)
		}

		for _, u := range users {
			if u.TelegramID == nil {
				continue
			}

			telegramID := *u.TelegramID

			chat, err := h.bot.GetChat(ctx, &telego.GetChatParams{ChatID: tu.ID(telegramID)})
			if err != nil {
				return fmt.Errorf("retrieving chat via bot: %w", err)
			}
			draftKb = append(draftKb, tu.InlineKeyboardRow(
				tu.
					InlineKeyboardButton(chat.LastName).
					WithCallbackData(fmt.Sprintf("%s:%d", ManageUsersPrefix, telegramID)),
			))
		}

		response.Text = "Доступные пользователи"
		draftKb = setPastDueKb
	case session.StateSetPastDueConfirm:
		_, err := h.subscriptionService.SetPeriodEnd(ctx, time.Now())
		if err != nil {
			return fmt.Errorf("trying to set period end: %w", err)
		}

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

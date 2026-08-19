package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/tgbot/session"
)

var ManageUsersPrefix = fmt.Sprintf("%s:manage_users", AdminHandlerPrefix)

var manageUsersActionsKb = [][]telego.InlineKeyboardButton{
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Изменить количество ключей").
			WithCallbackData(fmt.Sprintf("%s:edit_keys_amount", ManageUsersPrefix)),
	),
}

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
			fmt.Sprintf("%s", ManageUsersPrefix),
			"",
		)
		switch {
		case callbackData == "":
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateManageUsers
			})
		case strings.Contains(callbackData, ":u:"):
			userUUID, err := uuid.Parse(strings.ReplaceAll(callbackData, ":u:", ""))
			if err != nil {
				return fmt.Errorf("parsing user's UUID: %w", err)
			}

			h.sessions.Update(userID, func(s *session.Session) {
				s.EditUser.User.PublicID = userUUID
				s.State = session.StateManageUsersSelect
			})
		case callbackData == ":edit_keys_amount":
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateSetPastDueConfirm
			})
		case callbackData == ":edit_keys_amount":
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateManageUsersEditKeysAmount
			})
		}
	case update.Message != nil:
		userID = update.Message.From.ID
		response.ChatID = update.Message.Chat.ChatID()

		sess := h.sessions.Get(userID)
		switch sess.State {
		case session.StateManageUsers:
			newKeysLimit, err := strconv.Atoi(update.Message.Text)
			if err != nil {
				return fmt.Errorf("converting message with new keys amount to int: %w", err)
			}

			err = h.subscriptionService.SetKeysLimit(ctx, sess.EditUser.User.ID, newKeysLimit)
			if err != nil {
				return fmt.Errorf("setting new keys limit: %w", err)
			}

			response.Text = "Лимит ключей успешно обновлен!"
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateIdle
			})
		}
	}

	var draftKb [][]telego.InlineKeyboardButton

	sess := h.sessions.Get(userID)
	switch sess.State {
	case session.StateManageUsers:
		users, err := h.userService.ListAll(ctx)
		if err != nil {
			return fmt.Errorf("retrieving users: %w", err)
		}

		if len(users) == 0 {
			response.Text = "Нет доступных пользователей"

			break
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
					InlineKeyboardButton(chat.FirstName).
					WithCallbackData(fmt.Sprintf(
						"%s:u:%s",
						ManageUsersPrefix,
						u.PublicID.String(),
					)),
			))
		}

		response.Text = "Доступные пользователи"
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateManageUsersEditKeysAmount
		})
	case session.StateManageUsersSelect:
		subscription, err := h.subscriptionService.GetByPublicID(sess.EditUser.User.PublicID.String())
		if err != nil {
			return fmt.Errorf("retrieving user's subscription via its public ID: %w", err)
		}

		response.Text = fmt.Sprintf(
			"Пользователь %s\n\nЛимит ключей: %d",
			sess.EditUser.Username,
			subscription.Quantity,
		)
		draftKb = append(draftKb, manageUsersActionsKb...)
	case session.StateManageUsersEditKeysAmount:
		response.Text = "Введите новое количество ключей"
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateManageUsersEnterKeysAmount
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

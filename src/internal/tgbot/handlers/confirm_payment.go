package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/tgbot/session"
)

const ConfirmPaymentPrefix = "confirm_payment"

func (h *Handlers) HandleConfirmPayment(
	ctx *th.Context,
	update telego.Update,
) error {
	var response Response
	var userID int64
	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()
		userID = update.CallbackQuery.From.ID

		callbackData := strings.ReplaceAll(
			update.CallbackQuery.Data,
			fmt.Sprintf("%s", ConfirmPaymentPrefix),
			"",
		)
		switch callbackData {
		case "":
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateConfirmPayment
			})
		}
	case update.Message != nil:
		response.ChatID = update.Message.Chat.ChatID()
		userID = update.Message.From.ID
	}

	sess := h.sessions.Get(userID)
	switch sess.State {
	case session.StateConfirmPayment:
		err := h.subscriptionService.ClaimPayment(ctx, userID)
		if err != nil {
			return err
		}

		response.Text = "Запрос на подтверждение оплаты отправлен"
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateIdle
		})
	}

	if response.Text == "" {
		return nil
	}

	message := tu.Message(response.ChatID, response.Text).WithParseMode(telego.ModeMarkdownV2)
	response.ReplyMarkup = buildKbWithExitButton()
	message = message.WithReplyMarkup(response.ReplyMarkup)

	h.bot.SendMessage(
		ctx,
		message,
	)

	return nil
}

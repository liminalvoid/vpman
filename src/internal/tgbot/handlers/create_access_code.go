package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/lib"
	"vpnmanager/internal/tgbot/session"
)

const CreateInviteCodePrefix = "generate_invite"

func (h *Handlers) HandleGenerateAccessCode(
	ctx *th.Context,
	update telego.Update,
) error {
	var userID int64
	switch {
	case update.CallbackQuery != nil:
		userID = update.CallbackQuery.From.ID
	case update.Message != nil:
		userID = update.Message.From.ID
	}

	var response Response
	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		callbackData := strings.ReplaceAll(
			update.CallbackQuery.Data,
			fmt.Sprintf("%s:", AdminHandlerPrefix),
			"",
		)
		switch callbackData {
		case CreateInviteCodePrefix:
			accessCode, err := lib.GenerateAccessCode(8)
			if err != nil {
				return fmt.Errorf("trying generate invite code: %w", err)
			}

			err = h.accessCodeService.Create(ctx, accessCode)
			if err != nil {
				return fmt.Errorf("creating access code: %w", err)
			}

			response.Text = fmt.Sprintf(
				"Инвайт\\-код создан\\!\n\n```plain\n%s\n```",
				accessCode,
			)
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateIdle
			})
		}
	}

	if response.Text == "" {
		return nil
	}

	message := tu.Message(response.ChatID, response.Text).WithParseMode(telego.ModeMarkdownV2)
	response.ReplyMarkup = buildKbWithExitButton()
	message = message.WithReplyMarkup(response.ReplyMarkup)

	h.bot.SendMessage(ctx, message)

	return nil
}

package handlers

import (
	"context"
	"slices"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot/session"
)

type Response struct {
	Text        string
	ReplyMarkup telego.ReplyMarkup
	ChatID      telego.ChatID
}

type Handlers struct {
	bot           *telego.Bot
	sessions      *session.Sessions
	configService *service.ConfigService
	serverService *service.ServerService
}


func buildKbWithExitButton(rows ...[]telego.InlineKeyboardButton) *telego.InlineKeyboardMarkup {
	exitButton := [][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Назад").
				WithCallbackData("back"),
		),
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("⬅️ В меню").
				WithCallbackData("menu"),
		),
	}

	rows = append(rows, exitButton...)

	return tu.InlineKeyboard(rows...)
}

func NewHandlers(
	bot *telego.Bot,
	ps *service.ConfigService,
	ss *service.ServerService,
	sess *session.Sessions,
) *Handlers {
	return &Handlers{
		bot: bot,
		sessions: sess,
		configService: ps,
		serverService: ss,
	}
}

func (h *Handlers) InState(states ...session.WizardState) th.Predicate {
	return func(_ context.Context, update telego.Update) bool {
		if update.Message == nil || update.Message.From == nil {
			return false
		}

		userId := update.Message.From.ID
		sess := h.sessions.Get(userId)
		return slices.Contains(states, sess.State)
	}
}

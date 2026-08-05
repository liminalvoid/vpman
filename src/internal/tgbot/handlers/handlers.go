package handlers

import (
	"context"
	"slices"
	"sync"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/service"
)

type Response struct {
	Text        string
	ReplyMarkup telego.ReplyMarkup
	ChatID      telego.ChatID
}

type Handlers struct {
	bot           *telego.Bot
	sessions      *Sessions
	configService *service.ConfigService
	serverService *service.ServerService
}

type WizardState int

const (
	StateIdle WizardState = iota

	StateCreateConfig
	StateCreateConfigEnterName
	StateCreateConfigSelectLocation
	StateCreateConfigSelectProtocol
	StateCreateConfigConfirm

	StateEditConfig
	StateEditConfigSelect
	StateEditConfigEnterName
	StateEditConfigDelete
)

func buildKbWithExitButton(rows ...[]telego.InlineKeyboardButton) *telego.InlineKeyboardMarkup {
	exitButton := tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("⬅️ В меню").
			WithCallbackData("menu"),
	)

	rows = append(rows, exitButton)

	return tu.InlineKeyboard(rows...)
}

func NewHandlers(
	bot *telego.Bot,
	ps *service.ConfigService,
	ss *service.ServerService,
) *Handlers {
	return &Handlers{
		bot: bot,
		sessions: &Sessions{
			mu: sync.Mutex{},
			m:  make(map[int64]*Session),
		},
		configService: ps,
		serverService: ss,
	}
}

func (h *Handlers) InState(states ...WizardState) th.Predicate {
	return func(_ context.Context, update telego.Update) bool {
		if update.Message == nil || update.Message.From == nil {
			return false
		}

		return slices.Contains(states, h.sessions.Get(update.Message.From.ID).State)
	}
}

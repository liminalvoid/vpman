package handlers

import (
	"sync"

	"github.com/mymmrac/telego"

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
	peerService   *service.ConfigService
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
		peerService:   ps,
		serverService: ss,
	}
}

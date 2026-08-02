package handlers

import (
	"sync"

	"github.com/mymmrac/telego"

	"vpnmanager/internal/model"
	"vpnmanager/internal/service"
)

type Response struct {
	Text        string
	ReplyMarkup telego.ReplyMarkup
	ChatID      telego.ChatID
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

type Key struct {
	State       WizardState
	Name        string
	Server      *model.Server
	Protocol    string
	Location    string
	ContainerID string
	KeyID       string
}

type Handlers struct {
	bot           *telego.Bot
	keys          map[int64]Key
	mutex         *sync.RWMutex
	peerService   *service.ConfigService
	serverService *service.ServerService
}

func NewHandlers(
	bot *telego.Bot,
	ps *service.ConfigService,
	ss *service.ServerService,
) *Handlers {
	return &Handlers{
		bot:           bot,
		keys:          make(map[int64]Key),
		mutex:         &sync.RWMutex{},
		peerService:   ps,
		serverService: ss,
	}
}

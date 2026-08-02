package handlers

import (
	"sync"

	"github.com/mymmrac/telego"

	"vpnmanager/internal/model"
	"vpnmanager/internal/service"
)

type Key struct {
	State       GenerateConfigState
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

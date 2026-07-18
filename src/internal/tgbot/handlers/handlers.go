package handlers

import (
	"sync"

	"github.com/mymmrac/telego"

	"vpnmanager/internal/service"
)

type Key struct {
	State    GenerateConfigState
	Name     string
	Location string
	Protocol string
}

type Handlers struct {
	bot           *telego.Bot
	keys          map[int64]Key
	mutex         *sync.RWMutex
	peerService   *service.PeerService
	serverService *service.ServerService
}

func NewHandlers(
	bot *telego.Bot,
	ps *service.PeerService,
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

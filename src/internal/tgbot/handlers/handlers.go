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
	bot   *telego.Bot
	keys  map[int64]Key
	mutex *sync.RWMutex
	ps    *service.PeerService
}

func NewHandlers(
	bot *telego.Bot,
	ps *service.PeerService,
) *Handlers {
	return &Handlers{
		bot:   bot,
		keys:  make(map[int64]Key),
		mutex: &sync.RWMutex{},
		ps:    ps,
	}
}

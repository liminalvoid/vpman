package handlers

import (
	"sync"

	"github.com/mymmrac/telego"

	"vpnmanager/internal/configgen"
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
	cg    *configgen.ConfigGenerator
}

func NewHandlers(
	bot *telego.Bot,
	cg *configgen.ConfigGenerator,
) *Handlers {
	return &Handlers{
		bot:   bot,
		keys:  make(map[int64]Key),
		mutex: &sync.RWMutex{},
		cg:    cg,
	}
}

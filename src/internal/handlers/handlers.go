package handlers

import (
	"sync"

	"github.com/mymmrac/telego"
)

type Key struct {
	State    State
	Name     string
	Protocol string
}

type Handlers struct {
	bot   *telego.Bot
	keys  map[int64]Key
	mutex *sync.RWMutex
}

func NewHandlers(bot *telego.Bot) *Handlers {
	return &Handlers{
		bot:   bot,
		keys:  make(map[int64]Key),
		mutex: &sync.RWMutex{},
	}
}


package tgbot

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot/handlers"
	"vpnmanager/internal/tgbot/middleware"
)

type Config struct {
	Token   string
	AdminID int64
}

type Bot struct {
	handler *th.BotHandler
}

func New(
	ctx context.Context,
	cfg Config,
	ps *service.ConfigService,
	ss *service.ServerService,
) (*Bot, error) {
	// FIXME: Remove this and think how to handle logs in prod
	bot, err := telego.NewBot(cfg.Token, telego.WithDefaultDebugLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %s", err)
	}

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	bh, _ := th.NewBotHandler(bot, updates)
	h := handlers.NewHandlers(
		bot,
		ps,
		ss,
	)

	bh.Use(th.PanicRecovery())
	bh.Use(middleware.IsAdmin(bot, cfg.AdminID))

	bh.Handle(
		h.HandleStart,
		th.Or(
			th.CommandEqual("start"),
			th.CallbackDataContains("menu"),
		),
	)
	bh.Handle(
		h.HandleNewConfig,
		th.Or(
			th.CallbackDataContains("generate_key"),
			th.And(th.AnyMessage(), h.InState(handlers.StateCreateConfigEnterName)),
		),
	)
	bh.Handle(
		h.HandleManageKeys,
		th.Or(
			th.CallbackDataContains("manage_keys"),
			th.And(th.AnyMessage(), h.InState(handlers.StateEditConfigEnterName)),
		),
	)

	return &Bot{handler: bh}, nil
}

func (b *Bot) Start() error { return b.handler.Start() }
func (b *Bot) Stop()        { b.handler.Stop() }

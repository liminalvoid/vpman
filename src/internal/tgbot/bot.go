package tgbot

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot/handlers"
	"vpnmanager/internal/tgbot/middleware"
	"vpnmanager/internal/tgbot/session"
)

type Config struct {
	Token string
}

type Bot struct {
	handler *th.BotHandler
}

func New(
	ctx context.Context,
	cfg Config,
	cs *service.ConfigService,
	ss *service.ServerService,
	us *service.UserService,
	acs *service.AccessCodeService,
) (*Bot, error) {
	// FIXME: Remove this and think how to handle logs in prod
	bot, err := telego.NewBot(cfg.Token, telego.WithDefaultDebugLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %s", err)
	}

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	bh, _ := th.NewBotHandler(bot, updates)
	sessions := session.NewStore()
	h := handlers.NewHandlers(
		bot,
		sessions,
		cs,
		ss,
		acs,
	)

	bh.Use(th.PanicRecovery())
	bh.Use(middleware.IsAuthenticated(bot, us, acs))

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
			th.And(th.AnyMessage(), h.InState(session.StateCreateConfigEnterName)),
		),
	)
	bh.Handle(
		h.HandleManageKeys,
		th.Or(
			th.CallbackDataContains("manage_keys"),
			th.And(th.AnyMessage(), h.InState(session.StateEditConfigEnterName)),
		),
	)

	admin := bh.Group(th.Or(
		th.CommandPrefix(fmt.Sprintf("%s_", handlers.AdminHandlerPrefix)),
		th.CallbackDataPrefix(fmt.Sprintf("%s:", handlers.AdminHandlerPrefix)),
	))
	admin.Use(middleware.IsAdmin(bot, us))

	admin.Handle(
		h.HandleGenerateAccessCode,
		th.CallbackDataContains(
			fmt.Sprintf(
				"%s:%s",
				handlers.AdminHandlerPrefix,
				handlers.CreateInviteCodePrefix,
			),
		),
	)

	return &Bot{handler: bh}, nil
}

func (b *Bot) Start() error { return b.handler.Start() }
func (b *Bot) Stop()        { b.handler.Stop() }

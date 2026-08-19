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
	handler    *th.BotHandler
}

func New(
	ctx context.Context,
	tg *telego.Bot,
	cfg Config,
	cs *service.ConfigService,
	ss *service.ServerService,
	us *service.UserService,
	acs *service.AccessCodeService,
	sbs *service.SubscriptionService,
) (*Bot, error) {
	updates, _ := tg.UpdatesViaLongPolling(ctx, nil)

	bh, _ := th.NewBotHandler(tg, updates)
	sessions := session.NewStore()
	h := handlers.NewHandlers(
		tg,
		sessions,
		cs,
		ss,
		acs,
		sbs,
		us,
	)

	bh.Use(th.PanicRecovery())
	bh.Use(middleware.IsAuthenticated(tg, us, acs, sessions))

	bh.Handle(
		h.HandleStart,
		th.Or(
			th.CommandEqual("start"),
			th.CallbackDataContains("menu"),
			middleware.HasAuthState(middleware.AuthOK),
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
	admin.Use(middleware.IsAdmin(tg, us, sessions))

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
	admin.Handle(
		h.HandleSetPastDue,
		th.CallbackDataContains(
			handlers.SetPastDuePrefix,
		),
	)
	admin.Handle(
		h.HandleManageUsers,
		th.CallbackDataContains(
			handlers.ManageUsersPrefix,
		),
	)

	return &Bot{handler: bh}, nil
}

func (b *Bot) Start() error { return b.handler.Start() }
func (b *Bot) Stop()        { b.handler.Stop() }

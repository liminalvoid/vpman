package tgbot

import (
	"context"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"vpnmanager/internal/repository"
	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot/handlers"
	"vpnmanager/internal/tgbot/middleware"
	"vpnmanager/internal/tgbot/session"
)

type Bot struct {
	handler *th.BotHandler
}

func New(
	ctx context.Context,
	tg *telego.Bot,
	cs *service.ConfigService,
	ss *service.ServerService,
	us *service.UserService,
	sbs *service.SubscriptionService,
	acr *repository.AccessCode,
) (*Bot, error) {
	updates, _ := tg.UpdatesViaLongPolling(ctx, nil)

	bh, _ := th.NewBotHandler(tg, updates)
	sessions := session.NewStore()
	h := handlers.NewHandlers(
		tg,
		sessions,
		cs,
		ss,
		sbs,
		us,
		acr,
	)

	bh.Use(th.PanicRecovery())
	bh.Use(middleware.IsAuthenticated(tg, us, sessions))

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
	bh.Handle(
		h.HandleConfirmPayment,
		th.CallbackDataContains(handlers.ConfirmPaymentPrefix),
	)

	admin := bh.Group(th.Or(
		th.And(th.AnyMessage(), h.InState(
			session.StateManageUsersEnterKeysAmount,
			session.StateSetPastDueEnterMessage,
			session.StateSetPastDueEnterAmount,
		)),
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
		th.Or(
			th.And(
				th.AnyMessage(),
				h.InState(
					session.StateSetPastDueEnterMessage,
					session.StateSetPastDueEnterAmount,
				),
			),
			th.CallbackDataContains(handlers.SetPastDuePrefix),
		),
	)
	admin.Handle(
		h.HandleManageUsers,
		th.Or(
			th.And(th.AnyMessage(), h.InState(session.StateManageUsersEnterKeysAmount)),
			th.CallbackDataContains(handlers.ManageUsersPrefix),
		),
	)

	return &Bot{handler: bh}, nil
}

func (b *Bot) Start() error { return b.handler.Start() }
func (b *Bot) Stop()        { b.handler.Stop() }

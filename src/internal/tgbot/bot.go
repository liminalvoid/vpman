package tgbot

import (
	"context"
	"fmt"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"vpnmanager/internal/configgen"
	"vpnmanager/internal/executor"
	"vpnmanager/internal/tgbot/handlers"
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
	ex *executor.ContainerExecutor,
	cg *configgen.ConfigGenerator,
) (*Bot, error) {
	// FIXME: Remove this and think how to handle logs in prod
	bot, err := telego.NewBot(cfg.Token, telego.WithDefaultDebugLogger())
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %s", err)
	}

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	bh, _ := th.NewBotHandler(bot, updates)
	h := handlers.NewHandlers(bot, cg, ex)

	bh.Use(th.PanicRecovery())
	bh.Use(
		func(ctx *th.Context, update telego.Update) error {
			userId := int64(0)
			userName := ""

			switch {
			case update.Message != nil:
				userId = update.Message.From.ID
				userName = update.Message.From.FirstName
			case update.CallbackQuery != nil:
				userId = update.CallbackQuery.From.ID
				userName = update.CallbackQuery.From.FirstName
			}

			if userId != cfg.AdminID {
				bot.Logger().Debugf(
					"%s is trying to access bot",
					userName,
				)

				return nil
			}

			bot.Logger().Debugf("Access granted")

			return ctx.Next(update)
		},
	)

	bh.Handle(
		h.HandleStart,
		th.CommandEqual("start"),
	)
	bh.Handle(
		h.HandleNewConfig,
		th.Or(
			th.CallbackDataContains("generate_key"),
			th.AnyMessage(),
		),
	)
	bh.Handle(
		h.HandleManageKeys,
		th.Or(
			th.CallbackDataContains("manage_keys"),
			th.AnyMessage(),
		),
	)

	return &Bot{handler: bh}, nil
}

func (b *Bot) Start() error { return b.handler.Start() }
func (b *Bot) Stop()        { b.handler.Stop() }

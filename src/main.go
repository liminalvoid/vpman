package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/liminalvoid/vpn-manager/internal/handlers"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	botToken := os.Getenv("TG_TOKEN")
	adminId, err := strconv.ParseInt(os.Getenv("ME_ID"), 10, 64)
	if err != nil {
		fmt.Println("The administrator ID wasn't parsed")
		os.Exit(1)
	}

	// FIXME: Remove this and think how to handle logs in prod
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	updates, _ := bot.UpdatesViaLongPolling(context.Background(), nil)

	bh, _ := th.NewBotHandler(bot, updates)
	h := handlers.NewHandlers(bot)

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

			if userId != adminId {
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
	defer func() { bh.Stop() }()

	err = bh.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

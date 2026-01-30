package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken := os.Getenv("TG_TOKEN")

	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		log.Fatal(err)
	}

	updates, _ := bot.UpdatesViaLongPolling(context.Background(), nil)

	bh, _ := th.NewBotHandler(bot, updates)

	startKb := tu.Keyboard(
		tu.KeyboardRow(
			tu.KeyboardButton("Сгенерировать ключ"),
		),
		tu.KeyboardRow(
			tu.KeyboardButton("Управление ключами"),
		),
	).
		WithResizeKeyboard().
		WithInputFieldPlaceholder("Меню")

	bh.Handle(
		func(ctx *th.Context, update telego.Update) error {
			bot.SendMessage(
				ctx,
				tu.Messagef(
					tu.ID(update.Message.Chat.ID),
					"Привет, %s\n\nКоличество ключей: 5/5\n\nДоступные локации:\n\n🇩🇪 Германия\n🇫🇷 Франция\n🇳🇱 Нидерланды",
					update.Message.From.FirstName,
				).
					WithReplyMarkup(startKb),
			)

			return nil
		},
		th.CommandEqual("start"),
	)

	defer func() { bh.Stop() }()

	bh.Start()
}

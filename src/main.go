package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type State int

const (
	StateDefault State = iota
	StateEnterKeyName
	StateSelectProtocol
	StateConfirm
)

type Key struct {
	State    State
	Name     string
	Protocol string
}

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

	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	updates, _ := bot.UpdatesViaLongPolling(context.Background(), nil)

	bh, _ := th.NewBotHandler(bot, updates)

	startKb := tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Сгенерировать ключ").
				WithCallbackData("wizard:generate_key"),
		),
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("Управление ключами").
				WithCallbackData("wizard:manage_keys"),
		),
	)

	bh.Use(th.PanicRecovery())
	bh.Use(
		func(ctx *th.Context, update telego.Update) error {
			switch {
			case update.Message != nil:
				userId := update.Message.From.ID
				userName := update.Message.From.FirstName

				if userId != adminId {
					bot.Logger().Debugf(
						"%s is trying to access bot",
						userName,
					)

					return nil
				}
			case update.CallbackQuery != nil:
				userId := update.CallbackQuery.From.ID
				userName := update.CallbackQuery.From.FirstName

				if userId != adminId {
					bot.Logger().Debugf(
						"%s is trying to access bot",
						userName,
					)

					return nil
				}
			}

			bot.Logger().Debugf("Access granted")

			return ctx.Next(update)
		},
	)

	keys := make(map[int64]Key)
	mutexLock := sync.RWMutex{}

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
	bh.Handle(
		func(ctx *th.Context, update telego.Update) error {
			userId := int64(0)

			switch {
			case update.CallbackQuery != nil:
				userId = update.CallbackQuery.From.ID
			case update.Message != nil:
				userId = update.Message.From.ID
			}

			if userId == 0 {
				return nil
			}

			mutexLock.RLock()
			key := keys[userId]
			mutexLock.RUnlock()

			switch {
			case update.CallbackQuery != nil:
				bot.AnswerCallbackQuery(
					ctx,
					tu.CallbackQuery(update.CallbackQuery.ID),
				)

				text := "Введи название ключа"
				key.State = StateEnterKeyName

				bot.SendMessage(
					ctx,
					tu.Message(
						update.
							CallbackQuery.
							Message.
							GetChat().
							ChatID(),
						text,
					),
				)
			case update.Message != nil:
				text := ""

				switch key.State {
				case StateDefault:
				case StateEnterKeyName:
					key.Name = update.Message.Text
					key.State = StateSelectProtocol
					text = "Выбери протокол"
				case StateSelectProtocol:
					key.Protocol = update.Message.Text
					key.State = StateConfirm
					text = "Создать ключ?"
				case StateConfirm:
					if update.Message.Text == "Да" {
						text = "Ключ успешно создан"
					} else {
						key.State = StateDefault
					}
				}

				bot.SendMessage(ctx, tu.Message(update.Message.Chat.ChatID(), text))
			}

			mutexLock.Lock()
			keys[userId] = key
			mutexLock.Unlock()

			return nil
		},
		th.Or(
			th.CallbackDataEqual("wizard:generate_key"),
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

package main

import (
	"context"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	tgToken := os.Getenv("TG_TOKEN")
	options := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
		bot.WithCallbackQueryDataHandler("button", bot.MatchTypePrefix, callbackHandler),
	}

	b, err := bot.New(tgToken, options...)
	if err != nil {
		panic(err)
	}

	b.Start(ctx)
}

func callbackHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Germany, XRay", CallbackData: "button_key_001"},
			},
			{
				{Text: "Netherlands, AWG", CallbackData: "button_key_002"},
			},
			{
				{Text: "Germany, AWG", CallbackData: "button_key_003"},
			},
		},
	}

	if update.CallbackQuery.Data == "button_keys" {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.CallbackQuery.Message.Message.Chat.ID,
			Text:        "Сгенерированные ключи",
			ReplyMarkup: kb,
		})
	}

	b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		ShowAlert:       false,
	})
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	senderId := update.Message.From.ID
	meId, err := strconv.ParseInt(os.Getenv("ME_ID"), 10, 64)
	if err != nil {
		panic(err)
	}

	// if not me don't do anything ;)
	if senderId != meId {
		return
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Ключи", CallbackData: "button_keys"},
			},
			{
				{Text: "Управление подписками", CallbackData: "button_subscriptions"},
			},
		},
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Привет!\n",
		ReplyMarkup: kb,
	})
}

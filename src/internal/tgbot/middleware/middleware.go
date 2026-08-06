package middleware

import (
	"vpnmanager/internal/tgbot/session"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func IsAdmin(
	bot *telego.Bot,
	adminId int64,
) func(ctx *th.Context, update telego.Update) error {
	return func(ctx *th.Context, update telego.Update) error {
		var userId int64
		var userName string

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
	}
}

func GoBack(bot *telego.Bot, sessions *session.Sessions) th.Handler {
	return func(ctx *th.Context, update telego.Update) error {
		cq := update.CallbackQuery
		if cq == nil || cq.Data != "back" {
			return ctx.Next(update)
		}

		return ctx.Next(update)
	}
}

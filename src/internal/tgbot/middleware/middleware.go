package middleware

import (
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"vpnmanager/internal/service"
)

func IsAdmin(
	bot *telego.Bot,
	userService *service.UserService,
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

		bot.Logger().Debugf(
			"%s is trying to access bot",
			userName,
		)

		user, err := userService.GetByTelegramID(ctx, userId)
		if err != nil {
			return fmt.Errorf("querying user by Telegram ID: %w", err)
		}

		if !user.IsAdmin {
			return nil
		}

		bot.Logger().Debugf("Access granted")

		return ctx.Next(update)
	}
}

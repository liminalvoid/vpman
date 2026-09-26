package middleware

import (
	"errors"
	"fmt"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"vpnmanager/internal/model"
	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot/session"
)

func IsAuthenticated(
	bot *telego.Bot,
	userService *service.UserService,
	sessions *session.Sessions,
) func(ctx *th.Context, update telego.Update) error {
	return func(ctx *th.Context, update telego.Update) error {
		var userID int64
		var userName string
		var accessCode string

		switch {
		case update.Message != nil:
			userID = update.Message.From.ID
			userName = update.Message.From.FirstName

			accessCode = update.Message.Text
		case update.CallbackQuery != nil:
			userID = update.CallbackQuery.From.ID
			userName = update.CallbackQuery.From.FirstName
		}

		bot.Logger().Debugf(
			"%s is trying to access bot",
			userName,
		)

		user, err := userService.GetByTelegramID(ctx, userID)
		if errors.Is(err, model.ErrUserNotFound) {
			if accessCode == "" || len(accessCode) > 8 {
				return ctx.Next(update)
			}

			err := userService.CreateWithAccessCode(ctx, userID, accessCode)
			if errors.Is(err, model.ErrAccessCodeNotFound) {
				return fmt.Errorf("no such access code %s: %w", accessCode, err)
			} else if err != nil {
				return fmt.Errorf("removing access code and creating user: %w", err)
			}

			return ctx.WithValue(authStateKey{}, AuthOK).Next(update)
		} else if err != nil {
			return fmt.Errorf("querying user by Telegram ID: %w", err)
		}

		if user.IsAdmin {
			sessions.Update(userID, func(s *session.Session) {
				s.IsAdmin = true
			})
		}

		bot.Logger().Debugf("Access granted")

		return ctx.Next(update)
	}
}

var ErrUserNotAdmin = errors.New("user is not admin")

func IsAdmin(
	bot *telego.Bot,
	userService *service.UserService,
	sessions *session.Sessions,
) func(ctx *th.Context, update telego.Update) error {
	return func(ctx *th.Context, update telego.Update) error {
		var userID int64
		var userName string

		switch {
		case update.Message != nil:
			userID = update.Message.From.ID
			userName = update.Message.From.FirstName
		case update.CallbackQuery != nil:
			userID = update.CallbackQuery.From.ID
			userName = update.CallbackQuery.From.FirstName
		}

		bot.Logger().Debugf(
			"%s is trying to access protected handlers",
			userName,
		)

		if sess := sessions.Get(userID); !sess.IsAdmin {
			return ErrUserNotAdmin
		}
		
		bot.Logger().Debugf(
			"Access granted for admin-protected routes for %s",
			userName,
		)

		return ctx.Next(update)
	}
}

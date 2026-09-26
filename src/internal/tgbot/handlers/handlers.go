package handlers

import (
	"context"
	"slices"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/repository"
	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot/session"
)

type Response struct {
	Text        string
	ReplyMarkup telego.ReplyMarkup
	ChatID      telego.ChatID
}

type Handlers struct {
	bot                  *telego.Bot
	sessions             *session.Sessions
	configService        *service.ConfigService
	serverService        *service.ServerService
	subscriptionService  *service.SubscriptionService
	userService          *service.UserService
	accessCodeRepository *repository.AccessCode
}

const AdminHandlerPrefix = "admin"

var markdownV2Escaper = strings.NewReplacer(
	`\`, `\\`, "_", `\_`, "*", `\*`, "[", `\[`, "]", `\]`, "(", `\(`, ")", `\)`,
	"~", `\~`, "`", "\\`", ">", `\>`, "#", `\#`, "+", `\+`, "-", `\-`, "=", `\=`,
	"|", `\|`, "{", `\{`, "}", `\}`, ".", `\.`, "!", `\!`,
)

// escapeMarkdownV2 escapes every character Telegram reserves in MarkdownV2,
// so arbitrary text (user input, error messages) can be sent as plain text.
func escapeMarkdownV2(s string) string {
	return markdownV2Escaper.Replace(s)
}

func buildKbWithExitButton(rows ...[]telego.InlineKeyboardButton) *telego.InlineKeyboardMarkup {
	exitButton := [][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton("⬅️ В меню").
				WithCallbackData("menu"),
		),
	}

	rows = append(rows, exitButton...)

	return tu.InlineKeyboard(rows...)
}

func NewHandlers(
	bot *telego.Bot,
	sess *session.Sessions,
	ps *service.ConfigService,
	ss *service.ServerService,
	sbs *service.SubscriptionService,
	us *service.UserService,
	acr *repository.AccessCode,
) *Handlers {
	return &Handlers{
		bot:                  bot,
		sessions:             sess,
		configService:        ps,
		serverService:        ss,
		subscriptionService:  sbs,
		userService:          us,
		accessCodeRepository: acr,
	}
}

func (h *Handlers) InState(states ...session.WizardState) th.Predicate {
	return func(_ context.Context, update telego.Update) bool {
		if update.Message == nil || update.Message.From == nil {
			return false
		}

		userId := update.Message.From.ID
		sess := h.sessions.Get(userId)
		return slices.Contains(states, sess.State)
	}
}

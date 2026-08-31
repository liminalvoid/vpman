package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/tgbot/session"
)

const ConfirmPaymentPrefix = "manage_keys"

func (h *Handlers) HandleConfirmPayment(ctx *th.Context, update telego.Update) error {}

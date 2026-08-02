package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

var manageKeyKb = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Изменить название").
			WithCallbackData("manage_keys:edit_name"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Удалить").
			WithCallbackData("manage_keys:delete"),
	),
)

func (h *Handlers) HandleManageKeys(ctx *th.Context, update telego.Update) error {
	var userID int64
	switch {
	case update.CallbackQuery != nil:
		userID = update.CallbackQuery.From.ID
	case update.Message != nil:
		userID = update.Message.From.ID
	default:
		return nil
	}

	response := Response{}

	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		switch {
		case update.CallbackQuery.Data == "wizard:manage_keys":
			h.sessions.Update(userID, func(s *Session) {
				s.State = StateEditConfig
			})

			configs, err := h.peerService.ListConfigs(ctx, response.ChatID.ID)
			if err != nil {
				return fmt.Errorf("querying configs: %w", err)
			}

			keysKbButtons := make([][]telego.InlineKeyboardButton, 0)
			for _, c := range configs {
				buttonText := fmt.Sprintf(
					"%s, %s",
					c.Name,
					c.Protocol,
				)

				keysKbButtons = append(
					keysKbButtons,
					tu.InlineKeyboardRow(
						tu.
							InlineKeyboardButton(buttonText).
							WithCallbackData(fmt.Sprintf("manage_keys:key-%s", c.PublicID.String())),
					),
				)
			}
			keysKb := tu.InlineKeyboard(keysKbButtons...)

			response.Text = "Текущие ключи"
			response.ReplyMarkup = keysKb
		case strings.Contains(update.CallbackQuery.Data, "manage_keys:key-"):
			keyID := strings.ReplaceAll(update.CallbackQuery.Data, "manage_keys:key-", "")
			h.sessions.Update(userID, func(s *Session) {
				s.State = StateEditConfigSelect
			})

			cfg, err := h.peerService.GetByPublicID(ctx, keyID)
			if err != nil {
				return fmt.Errorf("retrieving configs: %w", err)
			}
			
			response.Text = fmt.Sprintf(
				"%s\n\nЛокация: %s\nПротокол: %s\nСтатус: %s",
				cfg.Name,
				convertCountryCodeToFlagEmoji(cfg.Location),
				cfg.Protocol,
				cfg.Status,
			)
			response.ReplyMarkup = manageKeyKb
		case strings.Contains(update.CallbackQuery.Data, "edit_name"):
			h.sessions.Update(userID, func(s *Session) {
				s.State = StateEditConfigEnterName
			})
			response.Text = "Введите новое название"
		case strings.Contains(update.CallbackQuery.Data, "delete"):
			h.sessions.Update(userID, func(s *Session) {
				s.State = StateEditConfigDelete
			})
			response.Text = "Конфиг удален"
		}
	case update.Message != nil:
		response.ChatID = update.Message.Chat.ChatID()
	}

	if response.Text == "" {
		return nil
	}

	message := tu.Message(response.ChatID, response.Text)

	if response.ReplyMarkup != nil {
		message = message.WithReplyMarkup(response.ReplyMarkup)
	}

	if _, err := h.bot.SendMessage(ctx, message); err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	return nil
}

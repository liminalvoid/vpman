package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/tgbot/session"
)

var manageKeyKb = [][]telego.InlineKeyboardButton{
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
}

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

	var draftKb [][]telego.InlineKeyboardButton
	response := Response{}

	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		callbackData := update.CallbackQuery.Data
		switch {
		case callbackData == "manage_keys":
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateEditConfig
			})
		case callbackData == "manage_keys:edit_name":
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateEditConfigEditName
			})
		case callbackData == "manage_keys:delete":
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateEditConfigDelete
			})
		case strings.Contains(callbackData, "manage_keys:key-"):
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateEditConfigSelect
			})
		}
	case update.Message != nil:
		response.ChatID = update.Message.Chat.ChatID()
		
		sess := h.sessions.Get(userID)
		switch sess.State {
		case session.StateEditConfigEnterName:
			newName := update.Message.Text
			err := h.configService.UpdateConfigName(ctx, sess.Edit.Config.PublicID.String(), newName)
			if err != nil {
				return err
			}

			response.Text = "Название конфига успешно обновлено"
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateIdle
			})
		}
	}

	sess := h.sessions.Get(userID)
	switch sess.State {
	case session.StateEditConfig:
		configs, err := h.configService.ListConfigs(ctx, response.ChatID.ID)
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

		response.Text = "Текущие ключи"
		draftKb = keysKbButtons
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateEditConfigSelect
		})
	case session.StateEditConfigSelect:
		keyID := strings.ReplaceAll(update.CallbackQuery.Data, "manage_keys:key-", "")
		cfg, err := h.configService.GetConfigByPublicID(ctx, keyID)
		if err != nil {
			return fmt.Errorf("retrieving configs: %w", err)
		}

		h.sessions.Update(userID, func(s *session.Session) {
			s.Edit.Config = cfg
		})

		response.Text = fmt.Sprintf(
			"%s\n\nЛокация: %s\nПротокол: %s\nСтатус: %s",
			cfg.Name,
			convertCountryCodeToFlagEmoji(cfg.Location),
			cfg.Protocol,
			cfg.Status,
		)
		draftKb = manageKeyKb
	case session.StateEditConfigEditName:
		response.Text = "Введите новое название"
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateEditConfigEnterName
		})
	case session.StateEditConfigDelete:
		sess := h.sessions.Get(userID)
		serverPublicID := sess.Edit.Config.ServerPublicID.String()
		server, err := h.serverService.GetByPublicID(
			ctx,
			serverPublicID,
		)
		if err != nil {
			return fmt.Errorf("retrieving server %q: %w", serverPublicID, err)
		}

		err = h.configService.RevokeConfig(ctx, server, sess.Edit.Config)
		if err != nil {
			return fmt.Errorf(
				"revoking config %q: %w",
				sess.Edit.Config.PublicID.String(),
				err,
			)
		}

		response.Text = "Конфиг удален"
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateIdle
		})
	}

	if response.Text == "" {
		return nil
	}

	message := tu.Message(response.ChatID, response.Text)
	response.ReplyMarkup = buildKbWithExitButton(draftKb...)
	message = message.WithReplyMarkup(response.ReplyMarkup)

	if _, err := h.bot.SendMessage(ctx, message); err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	return nil
}

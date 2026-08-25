package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"vpnmanager/internal/service"
	"vpnmanager/internal/tgbot/session"
)

const GenerateNewConfigPrefix = "generate_key"

var submitActionKb = [][]telego.InlineKeyboardButton{
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Да").
			WithCallbackData(fmt.Sprintf("%s:yes", GenerateNewConfigPrefix)),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Изменить").
			WithCallbackData(fmt.Sprintf("%s:edit", GenerateNewConfigPrefix)),
	),
}

var finalMenuKb = [][]telego.InlineKeyboardButton{
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Сгенерировать ещё раз").
			WithCallbackData(GenerateNewConfigPrefix),
	),
}

var selectConfigTypeKb = [][]telego.InlineKeyboardButton{
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Нативный формат").
			WithCallbackData(fmt.Sprintf("%s:native", GenerateNewConfigPrefix)),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Для AmneziaVPN").
			WithCallbackData(fmt.Sprintf("%s:amnezia", GenerateNewConfigPrefix)),
	),
}

func convertCountryCodeToFlagEmoji(code string) string {
	if len(code) != 2 {
		return ""
	}

	// Regional Indicator Symbol
	risOffset := 127397
	code = strings.ToUpper(code)
	r := []rune(code)

	return string(r[0]+rune(risOffset)) + string(r[1]+rune(risOffset))
}

func (h *Handlers) HandleNewConfig(ctx *th.Context, update telego.Update) error {
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

	servers, err := h.serverService.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("retrieving servers: %w", err)
	}

	sess := h.sessions.Get(userID)
	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		callbackData := update.CallbackQuery.Data
		switch {
		case callbackData == GenerateNewConfigPrefix:
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateCreateConfig
			})
		case callbackData == fmt.Sprintf("%s:yes", GenerateNewConfigPrefix):
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateCreateConfigConfirm
			})
		case callbackData == fmt.Sprintf("%s:edit", GenerateNewConfigPrefix):
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateCreateConfigEdit
			})
		case strings.Contains(callbackData, fmt.Sprintf("%s:protocol-", GenerateNewConfigPrefix)):
			sess = h.sessions.Update(userID, func(s *session.Session) {
				s.Create.Protocol = strings.ReplaceAll(update.CallbackQuery.Data, fmt.Sprintf("%s:protocol-", GenerateNewConfigPrefix), "")
			})
		case callbackData == fmt.Sprintf("%s:native", GenerateNewConfigPrefix):
			h.sessions.Update(userID, func(s *session.Session) {
				s.Create.Format = "native"
				s.State = session.StateCreateConfigPreview
			})
		case callbackData == fmt.Sprintf("%s:amnezia", GenerateNewConfigPrefix):
			h.sessions.Update(userID, func(s *session.Session) {
				s.Create.Format = "amnezia"
				s.State = session.StateCreateConfigPreview
			})
		}
	case update.Message != nil:
		response.ChatID = update.Message.Chat.ChatID()

		switch sess.State {
		case session.StateCreateConfigEnterName:
			configName := update.Message.Text

			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateCreateConfigSelectLocation
				s.Create.Name = configName
			})
		}
	}

	var availableProtocols map[string]struct{}
	sess = h.sessions.Get(userID)
	switch sess.State {
	case session.StateCreateConfig, session.StateCreateConfigEdit:
		if len(servers) == 0 {
			response.Text = "Нет доступных серверов"
			draftKb = finalMenuKb
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateIdle
			})

			break
		}

		isLimitSurpassed, err := h.subscriptionService.IsLimitSurpassedByTelegramID(ctx, userID)
		if err != nil || isLimitSurpassed == nil {
			return fmt.Errorf("retrieving config limit overrun: %w", err)
		}

		if *isLimitSurpassed {
			response.Text = "Не удалось создать конфиг: превышен допустимый лимит ключей"

			break
		}

		response.Text = "Введи название ключа"
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateCreateConfigEnterName
		})
	case session.StateCreateConfigSelectLocation:
		// Generate keyboard with available servers
		var serversKbRows [][]telego.InlineKeyboardButton
		for _, s := range servers {
			flagEmoji := convertCountryCodeToFlagEmoji(s.Location)
			keyTitle := fmt.Sprintf("%s %s | %s", flagEmoji, s.LocationCountry, s.LocationCity)
			callbackData := fmt.Sprintf("%s:server-%s", GenerateNewConfigPrefix, s.PublicID.String())

			serversKbRows = append(serversKbRows, tu.InlineKeyboardRow(
				tu.
					InlineKeyboardButton(keyTitle).
					WithCallbackData(callbackData),
			))
		}

		draftKb = serversKbRows
		response.Text = "Выбери локацию"
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateCreateConfigSelectProtocol
		})
	case session.StateCreateConfigSelectProtocol:
		if update.CallbackQuery == nil {
			break
		}

		selectedServer := strings.ReplaceAll(update.CallbackQuery.Data, fmt.Sprintf("%s:server-", GenerateNewConfigPrefix), "")
		for _, serv := range servers {
			if serv.PublicID.String() == selectedServer {
				sess = h.sessions.Update(userID, func(s *session.Session) {
					s.Create.Server = &serv
				})

				break
			}
		}

		if sess.Create.Server == nil {
			return fmt.Errorf("unknown server %q", selectedServer)
		}

		availableProtocols, err = h.configService.GetAvailableProtocols(ctx, sess.Create.Server)
		if err != nil {
			return fmt.Errorf("getting available protocols: %w", err)
		}

		if len(availableProtocols) == 0 {
			response.Text = "Нет доступных протоколов"
			draftKb = finalMenuKb
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateIdle
			})

			break
		}

		// Generate keyboard with available protocols
		var protocolsKbRows [][]telego.InlineKeyboardButton
		for name := range availableProtocols {
			callbackData := fmt.Sprintf("%s:protocol-%s", GenerateNewConfigPrefix, name)

			protocolsKbRows = append(protocolsKbRows, tu.InlineKeyboardRow(
				tu.
					InlineKeyboardButton(name).
					WithCallbackData(callbackData),
			))
		}

		response.Text = "Выбери протокол"
		draftKb = protocolsKbRows
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateCreateConfigSelectFormat
		})
	case session.StateCreateConfigSelectFormat:
		response.Text = "Выбери формат"
		draftKb = selectConfigTypeKb
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateCreateConfigPreview
		})
	case session.StateCreateConfigPreview:

		response.Text = fmt.Sprintf(
			"Название: %s\nПротокол: %s\nЛокация: %s\n\nСоздать ключ?",
			sess.Create.Name,
			sess.Create.Protocol,
			sess.Create.Server.Location,
		)
		draftKb = submitActionKb
	case session.StateCreateConfigConfirm:
		config, err := h.configService.CreateConfig(
			ctx,
			update.CallbackQuery.From.ID,
			sess.Create.Server,
			sess.Create.Name,
			sess.Create.Protocol,
			sess.Create.Protocol,
			sess.Create.Format,
		)
		if err != nil {
			if _, ok := errors.AsType[*service.LimitError](err); ok {
				response.Text = "Не удалось создать конфиг: превышен допустимый лимит ключей"
			} else {
				// TODO: Remove error sent to the user
				response.Text = fmt.Sprintf("Не удалось создать конфиг: %s", err)
			}

			draftKb = finalMenuKb
			h.sessions.Update(userID, func(s *session.Session) {
				s.State = session.StateIdle
			})

			break
		}

		response.Text = fmt.Sprintf("Ключ успешно создан\\!\n\n```plain\n%s\n```", config)
		draftKb = finalMenuKb
		h.sessions.Update(userID, func(s *session.Session) {
			s.State = session.StateIdle
		})
	}

	if response.Text == "" {
		return nil
	}

	message := tu.Message(response.ChatID, response.Text).WithParseMode(telego.ModeMarkdownV2)
	response.ReplyMarkup = buildKbWithExitButton(draftKb...)
	message = message.WithReplyMarkup(response.ReplyMarkup)

	h.bot.SendMessage(
		ctx,
		message,
	)

	return nil
}

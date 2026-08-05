package handlers

import (
	"fmt"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

var submitActionKb = [][]telego.InlineKeyboardButton{
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Да").
			WithCallbackData("generate_key:yes"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Изменить").
			WithCallbackData("generate_key:edit"),
	),
}

var finalMenuKb = [][]telego.InlineKeyboardButton{
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Сгенерировать ещё раз").
			WithCallbackData("wizard:generate_key"),
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
	var userId int64
	switch {
	case update.CallbackQuery != nil:
		userId = update.CallbackQuery.From.ID
	case update.Message != nil:
		userId = update.Message.From.ID
	default:
		return nil
	}

	var draftKb [][]telego.InlineKeyboardButton
	response := Response{}

	// Generate keyboard with available servers
	var serversKbRows [][]telego.InlineKeyboardButton
	servers, err := h.serverService.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("retrieving servers: %w", err)
	}
	for _, s := range servers {
		flagEmoji := convertCountryCodeToFlagEmoji(s.Location)
		keyTitle := fmt.Sprintf("%s %s | %s", flagEmoji, s.LocationCountry, s.LocationCity)
		callbackData := fmt.Sprintf("generate_key:server-%s", s.PublicID.String())

		serversKbRows = append(serversKbRows, tu.InlineKeyboardRow(
			tu.
				InlineKeyboardButton(keyTitle).
				WithCallbackData(callbackData),
		))
	}

	var availableProtocols map[string]struct{}

	switch {
	case update.CallbackQuery != nil:
		response.ChatID = update.CallbackQuery.Message.GetChat().ChatID()

		h.bot.AnswerCallbackQuery(
			ctx,
			tu.CallbackQuery(update.CallbackQuery.ID),
		)

		switch {
		case update.CallbackQuery.Data == "wizard:generate_key" || update.CallbackQuery.Data == "generate_key:edit":
			if len(servers) == 0 {
				h.sessions.Update(userId, func(s *Session) {
					s.State = StateCreateConfig
				})
				response.Text = "Нет доступных серверов"
				draftKb = finalMenuKb

				break
			}

			response.Text = "Введи название ключа"
			h.sessions.Update(userId, func(s *Session) {
				s.State = StateCreateConfigEnterName
			})
		case strings.Contains(update.CallbackQuery.Data, "server"):
			selectedServer := strings.ReplaceAll(update.CallbackQuery.Data, "generate_key:server-", "")

			for _, serv := range servers {
				if serv.PublicID.String() == selectedServer {
					h.sessions.Update(userId, func(s *Session) {
						s.Create.Server = &serv
					})

					break
				}
			}

			sess := h.sessions.Get(userId)

			availableProtocols, err = h.configService.GetAvailableProtocols(ctx, sess.Create.Server)
			if err != nil {
				return fmt.Errorf("getting available protocols: %w", err)
			}

			if len(availableProtocols) == 0 {
				h.sessions.Update(userId, func(s *Session) {
					s.State = StateCreateConfig
				})
				response.Text = "Нет доступных протоколов"
				draftKb = finalMenuKb

				break
			}

			// Generate keyboard with available protocols
			var protocolsKbRows [][]telego.InlineKeyboardButton
			for name := range availableProtocols {
				callbackData := fmt.Sprintf("generate_key:protocol-%s", name)

				protocolsKbRows = append(protocolsKbRows, tu.InlineKeyboardRow(
					tu.
						InlineKeyboardButton(name).
						WithCallbackData(callbackData),
				))
			}

			h.sessions.Update(userId, func(s *Session) {
				s.State = StateCreateConfigConfirm
			})
			response.Text = "Выбери протокол"
			draftKb = protocolsKbRows
		case strings.Contains(update.CallbackQuery.Data, "protocol"):
			h.sessions.Update(userId, func(s *Session) {
				s.Create.Protocol = strings.ReplaceAll(update.CallbackQuery.Data, "generate_key:protocol-", "")
				s.State = StateCreateConfigSelectProtocol
			})

			sess := h.sessions.Get(userId)

			response.Text = fmt.Sprintf(
				"Название: %s\nПротокол: %s\nЛокация: %s\n\nСоздать ключ?",
				sess.Create.Name,
				sess.Create.Protocol,
				sess.Create.Server.Location,
			)
			draftKb = submitActionKb
		case update.CallbackQuery.Data == "generate_key:yes":
			sess := h.sessions.Get(userId)
			config, err := h.configService.CreateConfig(
				ctx,
				update.CallbackQuery.From.ID,
				sess.Create.Server,
				sess.Create.Name,
				sess.Create.Protocol,
				sess.Create.Protocol,
			)
			if err != nil {
				// TODO: Remove error sent to the user
				response.Text = fmt.Sprintf("Не удалось создать конфиг: %s", err)
				draftKb = finalMenuKb

				h.sessions.Update(userId, func(s *Session) {
					s.State = StateCreateConfig
				})

				break
			}

			response.Text = fmt.Sprintf("Ключ успешно создан\\!\n\n```plain\n%s\n```", config)
			draftKb = finalMenuKb
			h.sessions.Update(userId, func(s *Session) {
				s.State = StateCreateConfig
			})
		}
	case update.Message != nil:
		response.ChatID = update.Message.Chat.ChatID()

		sess := h.sessions.Get(userId)
		switch sess.State {
		case StateCreateConfigEnterName:
			h.sessions.Update(userId, func(s *Session) {
				s.State = StateCreateConfigSelectProtocol
				s.Create.Name = update.Message.Text
			})

			response.Text = "Выбери локацию"
			draftKb = serversKbRows
		}
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

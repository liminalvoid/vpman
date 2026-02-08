package keyboards

import (
	tu "github.com/mymmrac/telego/telegoutil"
)

var Start = tu.InlineKeyboard(
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Сгенерировать ключ").
			WithCallbackData("wizard:generate_key"),
	),
	tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Управление ключами").
			WithCallbackData("wizard:manage_keys"),
	),
)

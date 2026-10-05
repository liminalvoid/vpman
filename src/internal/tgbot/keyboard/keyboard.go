package keyboard

import (
	"github.com/mymmrac/telego"
)

func paginate(
	keyboard []telego.InlineKeyboardButton,
	amountPerPage int,
	currentPage int,
) []telego.InlineKeyboardButton {
	if currentPage * amountPerPage > len(keyboard) {
		return []telego.InlineKeyboardButton{}
	}

	currPos := currentPage * amountPerPage + 1
	stop := currPos + amountPerPage

	return keyboard[currPos:stop]
}

package keyboard

import (
	"github.com/mymmrac/telego"
)

func Paginate(
	keyboard [][]telego.InlineKeyboardButton,
	amountPerPage int,
	currentPage int,
) [][]telego.InlineKeyboardButton {
	if currentPage * amountPerPage > len(keyboard) {
		return [][]telego.InlineKeyboardButton{}
	}

	currPos := currentPage * amountPerPage
	stop := currPos + amountPerPage

	return keyboard[currPos:stop]
}

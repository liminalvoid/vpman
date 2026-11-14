package keyboard

import (
	"fmt"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const RowsPerPage = 5

func BuildKbWithPagination(
	keyboard [][]telego.InlineKeyboardButton,
	prefix string,
	amountPerPage int,
	currentPage int,
) [][]telego.InlineKeyboardButton {
	currPos := currentPage * amountPerPage
	stop := min(currPos + amountPerPage, len(keyboard))

	resultKb := keyboard[currPos:stop]

	var navRow []telego.InlineKeyboardButton
	if currPos > 0 {
		navRow = append(
			navRow,
			tu.
				InlineKeyboardButton("⬅️ Назад").
				WithCallbackData(fmt.Sprintf("%s:page-%d", prefix, currentPage-1)),
		)
	}
	if stop < len(keyboard) {
		navRow = append(
			navRow,
			tu.
				InlineKeyboardButton("Далее ➡️").
				WithCallbackData(fmt.Sprintf("%s:page-%d", prefix, currentPage+1)),
		)
	}
	if len(navRow) > 0 {
		resultKb = append(resultKb, navRow)
	}

	return resultKb
}

func GetTotalPages(amount, rowsPerPage int) int {
	if amount <= 0 || rowsPerPage <= 0 {
		return 1
	}
	return (amount + rowsPerPage - 1) / rowsPerPage
}

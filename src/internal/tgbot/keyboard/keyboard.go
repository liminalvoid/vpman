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
	nextBtn := tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Далее ➡️").
			WithCallbackData(fmt.Sprintf("%s:page-%d", prefix, currentPage+1)),
	)
	backBtn := tu.InlineKeyboardRow(
		tu.
			InlineKeyboardButton("Назад ⬅️").
			WithCallbackData(fmt.Sprintf("%s:page-%d", prefix, currentPage-1)),
	)

	currPos := currentPage * amountPerPage
	stop := min(currPos + amountPerPage, len(keyboard))

	resultKb := keyboard[currPos:stop]
	if stop < len(keyboard) {
		resultKb = append(resultKb, nextBtn)
	}
	if currPos > 0 {
		resultKb = append(resultKb, backBtn)
	}

	return resultKb
}

func GetTotalPages(amount, rowsPerPage int) int {
	if amount <= 0 || rowsPerPage <= 0 {
		return 1
	}
	return (amount + rowsPerPage - 1) / rowsPerPage
}

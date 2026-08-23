package lib

import (
	"crypto/rand"
	"fmt"
)

const codeAlphabet = "0123456789ABCDEFGHIJKMNLOPQRSTUVWXYZ"

func GenerateAccessCode(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read((buf)); err != nil {
		return "", fmt.Errorf("generating code: %w", err)
	}

	code := make([]byte, length)
	for i, b := range buf {
		code[i] = codeAlphabet[int(b)%len(codeAlphabet)]
	}

	return string(code), nil
}

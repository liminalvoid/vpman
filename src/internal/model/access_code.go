package model

import (
	"time"
	"errors"
)

type AccessCode struct {
	ID        int64
	Code      string
	CreatedAt time.Time
}

var ErrAccessCodeNotFound = errors.New("access code not found")

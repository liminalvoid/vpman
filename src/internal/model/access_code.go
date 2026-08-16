package model

import "time"

type AccessCode struct {
	ID        int64
	Code      string
	CreatedAt time.Time
}

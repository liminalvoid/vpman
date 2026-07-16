package model

import (
	"github.com/google/uuid"
)

type User struct {
	ID           int64
	PublicID     uuid.UUID
	Username     string
	PasswordHash string
	IsAdmin      bool
}

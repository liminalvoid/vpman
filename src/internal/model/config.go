package model

import (
	"time"

	"github.com/google/uuid"
)

type ConfigStatus string

const (
	ConfigStatusActive  ConfigStatus = "active"
	ConfigStatusExpired ConfigStatus = "expired"
	ConfigStatusRevoked ConfigStatus = "revoked"
)

type Config struct {
	ID         int64
	PublicID   uuid.UUID
	Name       string
	UserID     int64
	ProtocolID int16
	Status     ConfigStatus
	Location   string
	CreatedAt  time.Time
}

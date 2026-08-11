package model

import (
	"time"

	"github.com/google/uuid"
)

type ConfigStatus string

const (
	ConfigStatusActive    ConfigStatus = "active"
	ConfigStatusSuspended ConfigStatus = "suspended"
)

type Config struct {
	ID             int64
	PublicID       uuid.UUID
	Name           string
	IPAddress      string
	ProtocolID     int16
	ServerID       int16
	SubscriptionID int64
	Status         ConfigStatus
	CreatedAt      time.Time
}

type ConfigListView struct {
	ID             int64
	PublicID       uuid.UUID
	Name           string
	IPAddress      string
	Protocol       string
	Location       string
	ServerPublicID uuid.UUID
	Status         ConfigStatus
	CreatedAt      time.Time
}

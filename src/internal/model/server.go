package model

import "github.com/google/uuid"

type Server struct {
	ID              int16
	PublicID        uuid.UUID
	PublicAddress   string
	InternalAddress string
	Location        string
	LocationCountry string
	LocationCity    string
}

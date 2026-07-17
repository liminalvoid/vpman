package model

type Server struct {
	ID              int16
	PublicAddress   string
	InternalAddress string
	Location        string
	LocationCountry string
	LocationCity    string
}

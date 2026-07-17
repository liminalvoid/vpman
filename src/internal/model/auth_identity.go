package model

type AuthIdentity struct {
	ID             int64
	UserID         int64
	ProviderID     int16
	ProviderUserID string
}

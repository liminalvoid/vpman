package handler

import "time"

type CreateConfigRequest struct {
	Name        string    `json:"name"         binding:"required"`
	Protocol    string    `json:"protocol"     binding:"required"`
	Status      string    `json:"status"       binding:"required,oneof=active expired revoked"`
	CountryCode string    `json:"country_code" binding:"required,iso3166_1_alpha2"`
	ExpiresAt   time.Time `json:"expires_at"   binding:"required"`
}

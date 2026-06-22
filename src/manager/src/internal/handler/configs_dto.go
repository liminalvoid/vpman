package handler

type CreateConfigRequest struct {
	Name     string `json:"name"         binding:"required,max=255"`
	Protocol string `json:"protocol"     binding:"required"`
	Status   string `json:"status"       binding:"required,oneof=active expired revoked"`
	Location string `json:"location"     binding:"required,iso3166_1_alpha2"`
}

type UpdateConfigRequest struct {
	Name     *string `json:"name"     binding:"omitempty,max=255"`
	Location *string `json:"location" binding:"omitempty,iso3166_1_alpha2"`
}

type ConfigResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
}

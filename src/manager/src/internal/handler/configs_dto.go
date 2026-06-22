package handler

type CreateConfigRequest struct {
	Name     string `json:"name"         binding:"required,max=255"`
	Protocol string `json:"protocol"     binding:"required"`
	Status   string `json:"status"       binding:"required,oneof=active expired revoked"`
	Location string `json:"location"     binding:"required,iso3166_1_alpha2"`
}

type ListConfigQuery struct {
	Status   string `form:"status"    binding:"omitempty,oneof=pending expired revoked"`
	Location string `form:"location"  binding:"omitempty,iso3166_1_alpha2"`
	Page     int    `form:"page"      binding:"omitempty,gte=1"`
	PageSize int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
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

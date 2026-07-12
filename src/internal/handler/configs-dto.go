package handler

type CreateConfigRequest struct {
	Name     string `json:"name"         binding:"required,max=255"`
	Protocol string `json:"protocol"     binding:"required"`
	Status   string `json:"status"       binding:"required,oneof=active"`
	Location string `json:"location"     binding:"required,iso3166_1_alpha2"`
}

type ConfigURI struct {
	PublicID string `uri:"id" binding:"required,uuid"`
}

type ListConfigQuery struct {
	Status   string `form:"status"    binding:"omitempty"`
	Location string `form:"location"  binding:"omitempty,iso3166_1_alpha2"`
	Page     int    `form:"page"      binding:"omitempty,gte=1"`
	PageSize int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
}

type UpdateConfigRequest struct {
	Name     *string `json:"name"     binding:"omitempty,max=255"`
}

type ConfigResponse struct {
	PublicID string `json:"id"`
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Location string `json:"location"`
}

type ConfigListResponse struct {
	Data       []ConfigResponse `json:"data"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalCount int              `json:"total_count"`
}

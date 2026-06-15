package handler

import (
	"context"

	"github.com/gin-gonic/gin"
)

type ConfigsService interface {
	Create(ctx context.Context, name, protocol string)
	GetById(ctx context.Context, id string)
	List(ctx context.Context, limit, offset int)
}

type ConfigsHandler struct {
	svc ConfigsService
}

func NewConfigsHandler(svc ConfigsService) *ConfigsHandler {
	return &ConfigsHandler{svc: svc}
}

func (h *ConfigsHandler) List(c *gin.Context) {

}

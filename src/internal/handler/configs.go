package handler

import (
	"github.com/gin-gonic/gin"
)

func create(c *gin.Context) {}
func get(c *gin.Context) {}
func update(c *gin.Context) {}
func delete(c *gin.Context) {}

func RegisterConfigsRoutes(rg *gin.RouterGroup) {
	configs := rg.Group("/configs")
	{
		configs.POST("", create)
		configs.GET("", get)
		configs.PATCH("", update)
		configs.DELETE("", delete)
	}
}

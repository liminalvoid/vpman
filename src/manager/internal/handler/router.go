package handler

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, configsHandler *ConfigsHandler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")	
	{
		configs := v1.Group("/configs")
		{
			configs.GET("", configsHandler.List)	
		}
	}
}

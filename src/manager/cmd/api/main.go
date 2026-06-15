package main

import (
	"manager/internal/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	handler.RegisterRoutes(router, handler.NewConfigsHandler())

	router.Run("localhost:8080")
}

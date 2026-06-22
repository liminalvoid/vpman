package main

import (
	"manager/internal/router"
)

func main() {
	router := router.SetupRouter()

	router.Run("localhost:8080")
}

package main

import (
	"manager/internal/handler"
)

func main() {
	router := handler.SetupRouter()

	router.Run("localhost:8080")
}

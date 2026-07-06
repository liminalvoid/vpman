package main

import (
	"manager/internal/router"
	"manager/internal/sshclient"
)

func main() {
	sshConfig := sshclient.Config{
		Host
	}

	r := router.SetupRouter()
	sc := sshclient.New()

	router.Run("localhost:8080")
}

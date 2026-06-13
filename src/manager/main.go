package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type config struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var configs = []config{
	{ID: "1", Name: "conf"},
}

func main() {
	router := gin.Default()

	router.GET("/configs", getConfigs)

	router.Run("localhost:8080")
}

func getConfigs(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, configs)
}

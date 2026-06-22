package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

func Error(c *gin.Context, status int, msg string) {
	c.JSON(status, ErrorResponse{Error: msg})
}

func ValidationError(c *gin.Context, fields map[string]string) {
	c.JSON(
		http.StatusBadRequest,
		ErrorResponse{
			Error:  "validation failed",
			Fields: fields,
		},
	)
}

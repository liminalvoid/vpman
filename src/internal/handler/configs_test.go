package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"manager/internal/router"
)

func TestCreateRoute(t *testing.T) {
	router := router.SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/configs", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}
func TestGetRoute(t *testing.T) {}
func TestUpdateRoute(t *testing.T) {}
func TestDeleteRoute(t *testing.T) {}

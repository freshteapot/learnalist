package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1GetRoot(c echo.Context) error {
	message := "1, 2, 3. Lets go!"
	response := api.HttpResponseMessage{
		Message: message,
	}

	return c.JSON(http.StatusOK, response)
}

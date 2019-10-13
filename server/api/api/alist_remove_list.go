package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1RemoveAlist(c echo.Context) error {
	r := c.Request()
	// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
	alist_uuid := strings.TrimPrefix(r.URL.Path, "/api/v1/alist/")

	user := c.Get("loggedInUser").(uuid.User)
	response := HttpResponseMessage{}

	err := m.Datastore.RemoveAlist(alist_uuid, user.Uuid)
	if err != nil {
		if err.Error() == i18n.SuccessAlistNotFound {
			response.Message = err.Error()
			return c.JSON(http.StatusNotFound, response)
		}

		if err.Error() == i18n.InputDeleteAlistOperationOwnerOnly {
			response.Message = err.Error()
			return c.JSON(http.StatusForbidden, response)
		}

		response.Message = i18n.InternalServerErrorDeleteAlist
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Remove from cache
	m.HugoHelper.Remove(alist_uuid)
	response.Message = fmt.Sprintf(i18n.ApiDeleteAlistSuccess, alist_uuid)
	return c.JSON(http.StatusOK, response)
}

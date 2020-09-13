package api

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1RemoveAlist(c echo.Context) error {
	alistUUID := c.Param("uuid")
	user := c.Get("loggedInUser").(uuid.User)
	response := api.HttpResponseMessage{}

	err := m.Datastore.RemoveAlist(alistUUID, user.Uuid)
	if err != nil {
		if err == i18n.ErrorListNotFound {
			response := api.HttpResponseMessage{
				Message: i18n.SuccessAlistNotFound,
			}
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
	m.HugoHelper.DeleteList(alistUUID)
	// TODO this might become a painful bottle neck
	m.HugoHelper.WriteListsByUser(user.Uuid, m.Datastore.GetAllListsByUser(user.Uuid))
	m.HugoHelper.WritePublicLists(m.Datastore.GetPublicLists())

	response.Message = fmt.Sprintf(i18n.ApiDeleteAlistSuccess, alistUUID)
	return c.JSON(http.StatusOK, response)
}

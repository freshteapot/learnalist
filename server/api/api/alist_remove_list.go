package api

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1RemoveAlist(c echo.Context) error {
	alistUUID := c.Param("uuid")
	user := c.Get("loggedInUser").(uuid.User)
	response := api.HTTPResponseMessage{}

	err := m.Datastore.RemoveAlist(alistUUID, user.Uuid)
	if err != nil {
		if err == i18n.ErrorListNotFound {
			response := api.HTTPResponseMessage{
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

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiListDelete,
		Data: event.EventList{
			UUID:     alistUUID,
			UserUUID: user.Uuid,
		},
	})

	response.Message = fmt.Sprintf(i18n.ApiDeleteAlistSuccess, alistUUID)
	return c.JSON(http.StatusOK, response)
}

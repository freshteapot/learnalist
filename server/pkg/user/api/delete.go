package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// @event.emit: event.ApiUserDelete
func (s UserService) V1DeleteUser(c echo.Context) error {
	response := api.HTTPResponseMessage{}
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid

	inputUUID := c.Param("uuid")
	if inputUUID != userUUID {
		response.Message = "You can only delete the user you are logged in with"
		return c.JSON(http.StatusForbidden, response)
	}

	err := s.userManagement.DeleteUser(userUUID)

	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"event":     event.UserDeleted,
			"error":     err,
			"user_uuid": userUUID,
		}).Error("problem")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiUserDelete,
		UUID: userUUID,
	})

	response.Message = "User has been removed"
	return c.JSON(http.StatusOK, response)
}

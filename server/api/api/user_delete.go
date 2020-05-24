package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/event"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func (m *Manager) V1DeleteUser(c echo.Context) error {
	logger := m.logger
	insights := m.insights
	response := HttpResponseMessage{}
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid

	inputUUID := c.Param("uuid")
	if inputUUID != userUUID {
		response.Message = "You can only delete the user you are logged in with"
		return c.JSON(http.StatusForbidden, response)
	}

	err := m.userManagement.DeleteUserFromDB(userUUID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":   event.UserDeleted,
			"context": "delete-from-db",
			"error":   err,
		}).Error("problem")
		response.Message = i18n.InternalServerErrorFunny
		return c.JSON(http.StatusInternalServerError, response)
	}

	insights.Event(logrus.Fields{
		"event":     event.UserDeleted,
		"user_uuid": userUUID,
	})

	response.Message = "User has been removed"
	return c.JSON(http.StatusOK, response)
}

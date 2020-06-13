package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func (m *Manager) V1DeleteUser(c echo.Context) error {
	logger := m.logger
	response := HttpResponseMessage{}
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid

	inputUUID := c.Param("uuid")
	if inputUUID != userUUID {
		response.Message = "You can only delete the user you are logged in with"
		return c.JSON(http.StatusForbidden, response)
	}

	err := m.userManagement.DeleteUser(userUUID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":     event.UserDeleted,
			"error":     err,
			"user_uuid": userUUID,
		}).Error("problem")
		response.Message = i18n.InternalServerErrorFunny
		return c.JSON(http.StatusInternalServerError, response)
	}

	m.HugoHelper.WritePublicLists(m.Datastore.GetPublicLists())
	response.Message = "User has been removed"
	return c.JSON(http.StatusOK, response)
}

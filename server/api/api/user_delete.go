package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1DeleteUser(c echo.Context) error {
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
		response.Message = err.Error()
		return c.JSON(http.StatusBadRequest, response)
	}

	response.Message = "User has been removed"
	return c.JSON(http.StatusOK, response)
}

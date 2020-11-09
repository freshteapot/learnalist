package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func (m *Manager) V1GetUserInfo(c echo.Context) error {
	logger := m.logger
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid

	inputUUID := c.Param("uuid")
	if inputUUID != userUUID {
		return c.JSON(http.StatusForbidden, api.HTTPResponseMessage{
			Message: "You can only get info for the user you are logged in with",
		})
	}

	b, err := m.UserManagement.GetInfo(userUUID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":     event.UserInfo,
			"api":       "V1GetUserInfo",
			"error":     err,
			"user_uuid": userUUID,
		}).Error("problem")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	var extra api.HTTPUserExtra
	err = json.Unmarshal(b, &extra)
	extra.CreatedVia = ""
	if extra.DisplayName == "" {
		extra.DisplayName = userUUID
	}

	return c.JSON(http.StatusOK, extra)
}

func (m *Manager) V1PatchUserInfo(c echo.Context) error {
	logger := m.logger
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid

	inputUUID := c.Param("uuid")
	if inputUUID != userUUID {
		return c.JSON(http.StatusForbidden, api.HTTPResponseMessage{
			Message: "You can only update info for the user you are logged in with",
		})
	}

	var input api.HTTPUserExtra
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	// On purpose dont let these be set.
	input.CreatedVia = ""

	b, _ := json.Marshal(input)
	err = m.UserManagement.SaveInfo(userUUID, b)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":     event.UserInfo,
			"api":       "V1PatchUserInfo",
			"error":     err,
			"user_uuid": userUUID,
		}).Error("problem")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	return c.NoContent(http.StatusOK)
}

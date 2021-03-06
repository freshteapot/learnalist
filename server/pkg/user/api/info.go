package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func (s UserService) V1GetUserInfo(c echo.Context) error {
	logger := s.logContext
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	userUUID := loggedInUser.Uuid

	inputUUID := c.Param("uuid")
	if inputUUID != userUUID {
		return c.JSON(http.StatusForbidden, api.HTTPResponseMessage{
			Message: i18n.UserInfoOnlyYourUUID,
		})
	}

	pref, err := s.userInfoRepo.Get(userUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			logger.WithFields(logrus.Fields{
				"event":     event.UserInfo,
				"api":       "V1GetUserInfo",
				"error":     err,
				"user_uuid": userUUID,
			}).Error("problem")
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}
	}

	pref.UserUUID = userUUID

	if pref.DailyReminder != nil {
		// this is to remove {}
		if (user.UserPreferenceDailyReminder{}) == *pref.DailyReminder {
			pref.DailyReminder = nil
		}
	}

	return c.JSON(http.StatusOK, pref)
}

func (s UserService) V1PatchUserInfo(c echo.Context) error {
	logger := s.logContext
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	userUUID := loggedInUser.Uuid

	inputUUID := c.Param("uuid")
	if inputUUID != userUUID {
		return c.JSON(http.StatusForbidden, api.HTTPResponseMessage{
			Message: "You can only update info for the user you are logged in with",
		})
	}

	var input openapi.HttpUserInfoInput
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	// TODO what to do here
	// On purpose dont let these be set.
	//input.CreatedVia = ""

	pref, err := s.userInfoRepo.Get(userUUID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":     event.UserInfo,
			"api":       "s.userInfoRepo.Get",
			"error":     err,
			"user_uuid": userUUID,
		}).Error("problem")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	pref.DisplayName = input.DisplayName
	if pref.DisplayName == "" {
		pref.DisplayName = userUUID
	}

	err = s.userInfoRepo.Save(userUUID, pref)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"event":     event.UserInfo,
			"api":       "V1PatchUserInfo",
			"error":     err,
			"user_uuid": userUUID,
		}).Error("problem")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	// TODO would need an event to get display name
	return c.NoContent(http.StatusOK)
}

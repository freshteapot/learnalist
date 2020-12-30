package app_settings

import (
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/user"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func NewService(userRepo user.ManagementStorage, log logrus.FieldLogger) AppSettingsService {
	s := AppSettingsService{
		userRepo:   userRepo,
		logContext: log,
	}
	return s
}

func (s AppSettingsService) SaveRemindV1(c echo.Context) error {
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	userUUID := loggedInUser.Uuid

	defer c.Request().Body.Close()
	var input openapi.AppSettingsRemindV1
	json.NewDecoder(c.Request().Body).Decode(&input)

	if input.SpacedRepetition.PushEnabled < 0 || input.SpacedRepetition.PushEnabled > 1 {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "push_enabled can only be 1 or 0",
		})
	}

	err := SaveRemindV1(s.userRepo, userUUID, input)

	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.userRepo.SaveInfo",
		}).Error("Issue with repo")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		UUID:   userUUID,
		Kind:   event.ApiAppSettingsRemindV1,
		Data:   input,
		Action: event.ActionUpsert,
	})
	return c.JSON(http.StatusOK, input)
}

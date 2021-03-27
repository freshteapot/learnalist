package remind

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type RemindService struct {
	userInfoRepo user.UserInfoRepository
	logContext   logrus.FieldLogger
}

func NewService(userInfoRepo user.UserInfoRepository, log logrus.FieldLogger) RemindService {
	s := RemindService{
		userInfoRepo: userInfoRepo,
		logContext:   log,
	}
	return s
}

func (s RemindService) GetDailySettings(c echo.Context) error {
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	userUUID := loggedInUser.Uuid
	appIdentifier := c.Param("appIdentifier")

	allowed := []string{apps.RemindV1, apps.PlankV1}
	if !utils.StringArrayContains(allowed, appIdentifier) {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "appIdentifier is not valid",
		})
	}

	pref, err := s.userInfoRepo.Get(userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if pref.DailyReminder == nil {
		return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
			Message: "Settings not found",
		})
	}

	var settings *openapi.RemindDailySettings
	switch appIdentifier {
	case apps.RemindV1:
		if pref.DailyReminder.RemindV1 != nil {
			settings = pref.DailyReminder.RemindV1
		}
	case apps.PlankV1:
		if pref.DailyReminder.PlankV1 != nil {
			settings = pref.DailyReminder.PlankV1
		}
	}

	if settings.AppIdentifier == "" {
		return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
			Message: "Settings not found",
		})
	}

	return c.JSON(http.StatusOK, settings)
}

func (s RemindService) DeleteDailySettings(c echo.Context) error {
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	userUUID := loggedInUser.Uuid
	appIdentifier := c.Param("appIdentifier")

	allowed := []string{apps.RemindV1, apps.PlankV1}
	if !utils.StringArrayContains(allowed, appIdentifier) {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "appIdentifier is not valid",
		})
	}

	pref, err := s.userInfoRepo.Get(userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	var settings *openapi.RemindDailySettings
	switch appIdentifier {
	case apps.RemindV1:
		if pref.DailyReminder.RemindV1 != nil {
			settings = pref.DailyReminder.RemindV1
			pref.DailyReminder.RemindV1 = nil
		}
	case apps.PlankV1:
		if pref.DailyReminder.PlankV1 != nil {
			settings = pref.DailyReminder.PlankV1
			pref.DailyReminder.PlankV1 = nil
		}
	}

	if settings == nil {
		return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
			Message: "Settings not found",
		})
	}

	err = s.userInfoRepo.Save(userUUID, pref)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		UUID:   userUUID,
		Kind:   EventApiRemindDailySettings,
		Data:   settings,
		Action: event.ActionDeleted,
	})

	return c.NoContent(http.StatusOK)
}

func (s RemindService) SetDailySettings(c echo.Context) error {
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	userUUID := loggedInUser.Uuid

	defer c.Request().Body.Close()
	var input openapi.RemindDailySettings
	json.NewDecoder(c.Request().Body).Decode(&input)

	// Validate app_identifier
	allowed := []string{apps.RemindV1, apps.PlankV1}
	if !utils.StringArrayContains(allowed, input.AppIdentifier) {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "appIdentifier is not valid",
		})
	}

	// Validate tz
	_, err := time.LoadLocation(input.Tz)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "tz is not valid",
		})
	}
	// Validate time_of_day
	input.TimeOfDay, err = ParseAndValidateTimeOfDay(input.TimeOfDay)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "time_of_day is not valid",
		})
	}
	// Validate medium
	allowed = []string{"push", "email"}
	if len(input.Medium) == 0 {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "medium is not valid",
		})
	}

	for _, medium := range input.Medium {
		if !utils.StringArrayContains(allowed, medium) {
			return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
				Message: "medium is not valid",
			})
		}
	}

	pref, err := s.userInfoRepo.Get(userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if pref.DailyReminder == nil {
		pref.DailyReminder = &user.UserPreferenceDailyReminder{}
	}

	switch input.AppIdentifier {
	case apps.RemindV1:
		pref.DailyReminder.RemindV1 = &input
	case apps.PlankV1:
		pref.DailyReminder.PlankV1 = &input
	}

	err = s.userInfoRepo.Save(userUUID, pref)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		UUID:   userUUID,
		Kind:   EventApiRemindDailySettings,
		Data:   input,
		Action: event.ActionUpsert,
	})

	return c.JSON(http.StatusOK, input)
}

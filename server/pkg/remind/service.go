package remind

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/app_settings"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/user"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type RemindService struct {
	userRepo   user.ManagementStorage
	logContext logrus.FieldLogger
}

func NewService(userRepo user.ManagementStorage, log logrus.FieldLogger) RemindService {
	s := RemindService{
		userRepo:   userRepo,
		logContext: log,
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

	response, err := s.getPreferences(userUUID, appIdentifier)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if response.AppIdentifier == "" {
		return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
			Message: "Settings not found",
		})
	}

	return c.JSON(http.StatusOK, response)
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

	response, err := s.getPreferences(userUUID, appIdentifier)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if response.AppIdentifier == "" {
		return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
			Message: "Settings not found",
		})
	}

	// This might break if we move from sqlite
	key := fmt.Sprintf(`%s.%s`, UserPreferenceKey, appIdentifier)
	err = s.userRepo.RemoveInfo(userUUID, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		UUID:   userUUID,
		Kind:   EventApiRemindDailySettings,
		Data:   response,
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
	err = ValidateTimeOfDay(input.TimeOfDay)
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

	info := user.UserPreference{
		DailyReminder: &user.UserPreferenceDailyReminder{},
	}

	switch input.AppIdentifier {
	case apps.RemindV1:
		info.DailyReminder.RemindV1 = &input
	case apps.PlankV1:
		info.DailyReminder.PlankV1 = &input
	}

	b, _ := json.Marshal(info)

	err = s.userRepo.SaveInfo(userUUID, b)
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

func (s RemindService) SetSpacedRepetition(c echo.Context) error {
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

	err := app_settings.SaveRemindV1(s.userRepo, userUUID, input)

	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.userRepo.SaveInfo",
		}).Error("Issue with repo")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		UUID:   userUUID,
		Kind:   EventApiRemindAppSettingsRemindV1,
		Data:   input,
		Action: event.ActionUpsert,
	})
	return c.JSON(http.StatusOK, input)
}

func (s RemindService) getPreferences(userUUID string, appIdentifier string) (openapi.RemindDailySettings, error) {
	var response openapi.RemindDailySettings
	b, err := s.userRepo.GetInfo(userUUID)
	if err != nil {
		return response, err
	}

	var pref user.UserPreference
	err = json.Unmarshal(b, &pref)
	if err != nil {
		return response, nil
	}

	if pref.DailyReminder == nil {
		return response, nil
	}

	switch appIdentifier {
	case apps.RemindV1:
		if pref.DailyReminder.RemindV1 != nil {
			response = *pref.DailyReminder.RemindV1
		}
	case apps.PlankV1:
		if pref.DailyReminder.PlankV1 != nil {
			response = *pref.DailyReminder.PlankV1
		}
	default:
		return response, errors.New("not supported")
	}

	return response, nil
}

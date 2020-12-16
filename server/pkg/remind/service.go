package remind

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
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

	//event.GetBus().Subscribe(event.TopicMonolog, "remind", s.monologSubscribe)
	return s
}

func (s RemindService) GetDailySettings(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	appIdentifier := c.Param("appIdentifier")

	allowed := []string{"remind:v1", "plank:v1"}
	if !utils.StringArrayContains(allowed, appIdentifier) {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "appIdentifier is not valid",
		})
	}

	b, err := s.userRepo.GetInfo(userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	var pref UserPreference
	json.Unmarshal(b, &pref)

	var response openapi.RemindDailySettings

	switch appIdentifier {
	case "remind:v1":
		if pref.DailyReminder.RemindV1 != nil {
			response = *pref.DailyReminder.RemindV1
		}
	case "plank:v1":
		if pref.DailyReminder.PlankV1 != nil {
			response = *pref.DailyReminder.PlankV1
		}
	}

	if response.AppIdentifier == "" {
		return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
			Message: "Settings not found",
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (s RemindService) DeleteDailySettings(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	appIdentifier := c.Param("appIdentifier")

	allowed := []string{"remind:v1", "plank:v1"}
	if !utils.StringArrayContains(allowed, appIdentifier) {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "appIdentifier is not valid",
		})
	}
	// Maybe I need user_preference code
	key := fmt.Sprintf(`daily_notifications."%s"`, appIdentifier)
	err := s.userRepo.RemoveInfo(userUUID, key)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	return c.NoContent(http.StatusOK)
}

func (s RemindService) SetDailySettings(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	appIdentifier := c.Param("appIdentifier")

	allowed := []string{"remind:v1", "plank:v1"}
	if !utils.StringArrayContains(allowed, appIdentifier) {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "appIdentifier is not valid",
		})
	}

	defer c.Request().Body.Close()

	var input openapi.RemindDailySettings
	json.NewDecoder(c.Request().Body).Decode(&input)

	if appIdentifier != input.AppIdentifier {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "appIdentifier is not valid",
		})
	}

	// TODO validate time of day
	// TODO validate tz?

	info := UserPreference{}
	switch appIdentifier {
	case "remind:v1":
		info.DailyReminder.RemindV1 = &input
	case "plank:v1":
		info.DailyReminder.PlankV1 = &input
	}

	b, _ := json.Marshal(info)

	s.userRepo.SaveInfo(userUUID, b)

	return c.JSON(http.StatusOK, input)
}

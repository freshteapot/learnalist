package mobile

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type MobileService struct {
	logContext logrus.FieldLogger
	repo       MobileRepository
}

func NewService(repo MobileRepository, log logrus.FieldLogger) MobileService {
	s := MobileService{
		repo:       repo,
		logContext: log,
	}

	event.GetBus().Subscribe(event.TopicMonolog, "mobile", s.OnEvent)
	return s
}

func (s MobileService) RegisterDevice(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	var registerInput openapi.HttpMobileRegisterInput
	json.Unmarshal(jsonBytes, &registerInput)

	if registerInput.Token == "" {
		response := api.HTTPResponseMessage{
			Message: "Token cant be empty",
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	status, err := s.repo.SaveDeviceInfo(userUUID, registerInput.Token)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":     err,
			"input":     string(jsonBytes),
			"user_uuid": userUUID,
		}).Error("Failed to register device")
		return c.JSON(http.StatusInternalServerError, api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		})
	}

	if status == http.StatusOK {
		return c.JSON(http.StatusOK, api.HTTPResponseMessage{
			Message: "Device registered",
		})
	}

	// Send a message to the log, that the device was registered
	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: EventMobileDeviceRegistered,
		Data: event.EventKV{
			UUID: userUUID,
			Data: DeviceInfo{
				Token:    registerInput.Token,
				UserUUID: userUUID,
			},
		},
	})

	return c.JSON(http.StatusOK, api.HTTPResponseMessage{
		Message: "Device registered",
	})
}
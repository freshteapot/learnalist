package mobile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type MobileService struct {
	logContext logrus.FieldLogger
}

func NewService(log logrus.FieldLogger) MobileService {
	s := MobileService{
		logContext: log,
	}

	event.GetBus().Subscribe("mobile", func(entry event.Eventlog) {
		fmt.Println("TODO")
	})
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

	// TODO register it
	s.logContext.WithFields(logrus.Fields{
		"input":     string(jsonBytes),
		"user_uuid": userUUID,
	}).Info("TODO save to db or push to event or something")

	event.GetBus().Publish(event.Eventlog{
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

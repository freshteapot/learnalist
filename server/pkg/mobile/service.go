package mobile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
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

	// TODO Update plank app, reject if "", today, assume plank
	if registerInput.AppIdentifier == "" {
		s.logContext.Warn("Fix humble plank to set plankV1 and then reject if empty")
		registerInput.AppIdentifier = apps.PlankV1
	}

	// TODO Update plank app, reject if "", today, assume plank
	if registerInput.AppIdentifier == "plank:v1" {
		s.logContext.Warn("Will go away after the mobile is updated")
		registerInput.AppIdentifier = "plank_v1"
	}

	allowed := []string{apps.PlankV1, apps.RemindV1}
	if !utils.StringArrayContains(allowed, registerInput.AppIdentifier) {
		response := api.HTTPResponseMessage{
			Message: fmt.Sprintf("App identifier is not supported: %s", strings.Join(allowed, ",")),
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	deviceInfo := openapi.MobileDeviceInfo{
		Token:         registerInput.Token,
		UserUuid:      userUUID,
		AppIdentifier: registerInput.AppIdentifier,
	}

	current, err := s.repo.GetDeviceInfoByToken(registerInput.Token)
	if err != nil {
		if err != ErrNotFound {
			return c.JSON(http.StatusInternalServerError, api.HTTPResponseMessage{
				Message: i18n.InternalServerErrorFunny,
			})
		}
	}
	// We fall thru on not found, meaning this should be empty
	if current.UserUuid != "" {
		response := api.HTTPResponseMessage{
			Message: "Token already registered",
		}

		if current.UserUuid != userUUID {
			return c.JSON(http.StatusUnprocessableEntity, response)
		}

		if current.AppIdentifier != registerInput.AppIdentifier {
			return c.JSON(http.StatusUnprocessableEntity, response)
		}

		return c.JSON(http.StatusOK, api.HTTPResponseMessage{
			Message: "Device registered",
		})
	}

	status, err := s.repo.SaveDeviceInfo(deviceInfo)
	if err != nil {
		if status == http.StatusUnprocessableEntity {
			return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
				Message: err.Error(),
			})
		}

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

	// TODO maybe add DeviceInfo to openapi
	// Send a message to the log, that the device was registered
	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind:   event.MobileDeviceRegistered,
		UUID:   userUUID,
		Data:   deviceInfo,
		Action: event.ActionUpsert,
	})

	return c.JSON(http.StatusOK, api.HTTPResponseMessage{
		Message: "Device registered",
	})
}

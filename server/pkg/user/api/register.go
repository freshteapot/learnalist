package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// V1PostRegister When a user is created it returns a 201.
// When a user is created with the same username and password it returns a 200.
// When a user is created with a username in the system it returns a 400.
func (s UserService) V1PostRegister(c echo.Context) error {
	if s.userRegisterKey != "" {
		registerKey := c.Request().Header.Get("x-user-register")
		if registerKey != s.userRegisterKey {
			return c.JSON(http.StatusForbidden, api.HTTPResponseMessage{
				Message: "User registration is locked down and requires key to add users",
			})
		}
	}

	var input openapi.HttpUserRegisterInput
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}
	cleanedUser := openapi.HttpUserRegisterInput{
		Username: input.Username,
		Password: input.Password,
	}

	// TODO Secure endpoint https://github.com/freshteapot/learnalist-api/issues/153
	extra := input.Extra
	if extra.CreatedVia != "plank.app.v1" {
		extra.CreatedVia = ""
	}

	// Validate display name
	if len(extra.DisplayName) > 20 {
		response := api.HTTPResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	cleanedUser, err = Validate(cleanedUser)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	hash := authenticate.HashIt(cleanedUser.Username, cleanedUser.Password)

	userUUID, err := s.userWithUsernameAndPassword.Lookup(cleanedUser.Username, hash)
	if err == nil {
		response := api.HTTPUserRegisterResponse{
			Uuid:     userUUID,
			Username: cleanedUser.Username,
		}
		return c.JSON(http.StatusOK, response)
	}

	aUser, err := s.userWithUsernameAndPassword.Register(cleanedUser.Username, hash)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Registering new user")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	pref := user.UserPreference{
		DisplayName: extra.DisplayName,
		CreatedVia:  "",
	}

	if input.Extra.GrantPublicListWriteAccess == "1" {
		pref.Acl.PublicListWrite = 1
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiUserRegister,
		Data: event.EventNewUser{
			UUID: aUser.UserUUID,
			Kind: event.KindUserRegisterUsername,
			Data: pref,
		},
	})

	response := api.HTTPUserRegisterResponse{
		Uuid:     aUser.UserUUID,
		Username: aUser.Username,
	}
	// Adding this, gives the event enough time to fire
	// Not quite sure if this a race in my tests or if this
	// Is a by product of the over engineering to bring events
	// into play
	time.Sleep(50 * time.Millisecond)
	return c.JSON(http.StatusCreated, response)
}

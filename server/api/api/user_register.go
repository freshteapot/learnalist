package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/user"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/labstack/echo/v4"
)

/*
When a user is created it returns a 201.
When a user is created with the same username and password it returns a 200.
When a user is created with a username in the system it returns a 400.

How to know its a throw away?
user_info
	json object
		throw_away:true,
		created_via: plank_app_v1,
		display_name: "chris",


event system to run hourly to keep a rolling window of users who have been active?
or create a

*/
func (m *Manager) V1PostRegister(c echo.Context) error {
	// If security header is missing, force creation to be a temporary user.
	// Or make all accounts temporary until we elevate them?
	var input api.HTTPUserRegisterInput
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}
	// Filter:
	//	created_via
	// 	displayName
	cleanedUser := api.HTTPUserRegisterInput{
		Username: input.Username,
		Password: input.Password,
	}

	extra := input.Extra
	extra.ThrowAway = "true"
	// TODO support register without throwaway
	// to allow a user to register without it being set as a throw away
	// check header for security access

	// TODO currently supported
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

	cleanedUser, err = user.Validate(cleanedUser)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	hash := authenticate.HashIt(cleanedUser.Username, cleanedUser.Password)

	userWithUsernameAndPassword := m.Datastore.UserWithUsernameAndPassword()
	userUUID, err := userWithUsernameAndPassword.Lookup(cleanedUser.Username, hash)
	if err == nil {
		response := api.HTTPUserRegisterResponse{
			Uuid:     userUUID,
			Username: cleanedUser.Username,
		}
		return c.JSON(http.StatusOK, response)
	}

	aUser, err := userWithUsernameAndPassword.Register(cleanedUser.Username, hash)
	if err != nil {
		// TODO Log this
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	extraB, _ := json.Marshal(extra)
	m.UserManagement.SaveInfo(aUser.UserUUID, extraB)

	event.GetBus().Publish(event.Eventlog{
		Kind: event.ApiUserRegister,
		Data: event.EventUser{
			UUID: aUser.UserUUID,
			Kind: event.KindUserRegisterUsername,
		},
	})

	response := api.HTTPUserRegisterResponse{
		Uuid:     aUser.UserUUID,
		Username: aUser.Username,
	}

	lists := make([]alist.ShortInfo, 0)
	m.HugoHelper.WriteListsByUser(aUser.UserUUID, lists)
	return c.JSON(http.StatusCreated, response)
}

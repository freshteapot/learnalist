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
	"github.com/labstack/echo/v4"
)

/*
When a user is created it returns a 201.
When a user is created with the same username and password it returns a 200.
When a user is created with a username in the system it returns a 400.
*/
func (m *Manager) V1PostRegister(c echo.Context) error {
	var input api.HttpUserRegisterInput
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response := api.HttpResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	cleanedUser := api.HttpUserRegisterInput{
		Username: input.Username,
		Password: input.Password,
	}

	cleanedUser, err = user.Validate(cleanedUser)
	if err != nil {

		response := api.HttpResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	hash := authenticate.HashIt(cleanedUser.Username, cleanedUser.Password)

	userWithUsernameAndPassword := m.Datastore.UserWithUsernameAndPassword()
	userUUID, err := userWithUsernameAndPassword.Lookup(cleanedUser.Username, hash)
	if err == nil {
		response := api.HttpUserRegisterResponse{
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

	response := api.HttpUserRegisterResponse{
		Uuid:     aUser.UserUUID,
		Username: aUser.Username,
	}

	lists := make([]alist.ShortInfo, 0)
	m.HugoHelper.WriteListsByUser(aUser.UserUUID, lists)
	return c.JSON(http.StatusCreated, response)
}

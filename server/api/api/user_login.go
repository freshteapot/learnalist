package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/user"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1PostLogin(c echo.Context) error {
	fmt.Println("Hello")
	var input api.HTTPUserRegisterInput
	response := api.HTTPResponseMessage{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response.Message = i18n.ApiUserLoginError
		return c.JSON(http.StatusBadRequest, response)
	}

	cleanedUser := api.HTTPUserRegisterInput{
		Username: input.Username,
		Password: input.Password,
	}

	cleanedUser, err = user.Validate(cleanedUser)
	if err != nil {
		response.Message = i18n.ApiUserLoginError
		return c.JSON(http.StatusBadRequest, response)
	}

	hash := authenticate.HashIt(cleanedUser.Username, cleanedUser.Password)

	userUUID, err := m.Datastore.UserWithUsernameAndPassword().Lookup(cleanedUser.Username, hash)
	if err != nil {
		response.Message = i18n.AclHttpAccessDeny
		return c.JSON(http.StatusForbidden, response)
	}

	session, err := m.Datastore.UserSession().NewSession(userUUID)
	if err != nil {
		response.Message = i18n.InternalServerErrorFunny
		return c.JSON(http.StatusInternalServerError, response)
	}

	cookie := authenticate.NewLoginCookie(session.Token)
	c.SetCookie(cookie)

	event.GetBus().Publish(event.Eventlog{
		Kind: event.ApiUserLogin,
		Data: event.EventUser{
			UUID: userUUID,
			Kind: event.KindUserLoginUsername,
		},
	})

	return c.JSON(http.StatusOK, &api.HTTPLoginResponse{
		Token:    session.Token,
		UserUUID: userUUID,
	})
}

package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1PostLogout(c echo.Context) error {
	var err error
	var input api.HTTPLogoutRequest
	defer c.Request().Body.Close()
	response := api.HttpResponseMessage{}
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err = json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response.Message = i18n.ApiUserLogoutError
		return c.JSON(http.StatusBadRequest, response)
	}

	switch input.Kind {
	case "token":
		break
	case "user":
		break
	default:
		response.Message = i18n.ApiUserLogoutError
		return c.JSON(http.StatusBadRequest, response)
	}

	if input.UserUUID == "" {
		response.Message = i18n.ApiUserLogoutError
		return c.JSON(http.StatusBadRequest, response)
	}

	if input.Token == "" {
		response.Message = i18n.ApiUserLogoutError
		return c.JSON(http.StatusBadRequest, response)
	}

	userSession := m.Datastore.UserSession()
	// Confirm the user is the user the token says it is
	userUUID, err := userSession.GetUserUUIDByToken(input.Token)
	if err != nil {
		if err != sql.ErrNoRows {
			response.Message = i18n.InternalServerErrorFunny
			return c.JSON(http.StatusInternalServerError, response)
		}

		response.Message = i18n.AclHttpAccessDeny
		return c.JSON(http.StatusForbidden, response)
	}

	if userUUID != input.UserUUID {
		response.Message = i18n.AclHttpAccessDeny
		return c.JSON(http.StatusForbidden, response)
	}

	eventKind := ""
	switch input.Kind {
	case "token":
		eventKind = event.KindUserLogoutSession
		err = userSession.RemoveSessionForUser(userUUID, input.Token)
		response.Message = fmt.Sprintf("Session %s, is now logged out", input.Token)
	case "user":
		eventKind = event.KindUserLogoutSessions
		err = userSession.RemoveSessionsForUser(userUUID)
		response.Message = fmt.Sprintf("All sessions have been logged out for user %s", userUUID)
	}

	if err != nil {
		response.Message = i18n.InternalServerErrorFunny
		return c.JSON(http.StatusInternalServerError, response)
	}

	event.GetBus().Publish(event.Eventlog{
		Kind: event.ApiUserLogout,
		Data: event.EventUser{
			UUID: userUUID,
			Kind: eventKind,
		},
	})

	return c.JSON(http.StatusOK, response)
}

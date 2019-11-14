package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/labstack/echo/v4"
)

type HTTPLogoutRequest struct {
	Kind     string `json:"kind"`
	UserUUID string `json:"user_uuid"`
	Token    string `json:"token"`
}

/*
200, 403
*/
func (m *Manager) V1PostLogout(c echo.Context) error {
	var err error
	var input HTTPLogoutRequest
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err = json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.ApiUserLogoutError,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	response := HttpResponseMessage{}
	switch input.Kind {
	case "token":
		fmt.Println("Logout a single token")
		err = m.Datastore.UserSession().RemoveSessionForUser(input.UserUUID, input.Token)
		response.Message = fmt.Sprintf("Session %s, is now logged out", input.Token)
	case "user":
		fmt.Println("Logout all sessions for the user")
		err = m.Datastore.UserSession().RemoveSessionsForUser(input.UserUUID)
		response.Message = fmt.Sprintf("All sessions have been logged out for user %s", input.UserUUID)
	default:
		response := HttpResponseMessage{
			Message: i18n.ApiUserLogoutError,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	if err != nil {
		response.Message = i18n.InternalServerErrorFunny
		return c.JSON(http.StatusInternalServerError, response)
	}

	return c.JSON(http.StatusOK, response)
}

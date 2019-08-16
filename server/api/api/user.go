package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/user"
	"github.com/labstack/echo/v4"
)

/*
When a user is created it returns a 201.
When a user is created with the same username and password it returns a 200.
When a user is created with a username in the system it returns a 400.
*/
func (m *Manager) V1PostRegister(c echo.Context) error {
	var input = &user.RegisterInput{}
	var cleanedUser user.RegisterInput

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	cleanedUser, err = user.Validate(*input)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	loginUser := authenticate.LoginUser{
		Username: cleanedUser.Username,
		Password: cleanedUser.Password,
	}

	statusCode := http.StatusCreated
	newUser, err := m.Datastore.GetUserByCredentials(loginUser)
	if err != nil {
		if err.Error() == i18n.DatabaseLookupNotFound {
			newUser, err = m.Datastore.InsertNewUser(loginUser)
			if err != nil {
				response := HttpResponseMessage{
					Message: err.Error(),
				}
				return c.JSON(http.StatusBadRequest, response)
			}
		}
	} else {
		statusCode = http.StatusOK
	}

	response := user.RegisterResponse{
		Uuid:     newUser.Uuid,
		Username: cleanedUser.Username,
	}
	return c.JSON(statusCode, response)
}

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
	var input HttpUserRegisterInput
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.ValidationUserRegister,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	cleanedUser := user.RegisterInput{
		Username: input.Username,
		Password: input.Password,
	}

	cleanedUser, err = user.Validate(cleanedUser)
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

	newUser, err := m.Datastore.InsertNewUser(loginUser)
	if err == nil {
		response := user.RegisterResponse{
			Uuid:     newUser.Uuid,
			Username: cleanedUser.Username,
		}
		return c.JSON(http.StatusCreated, response)
	}

	if err.Error() != i18n.UserInsertUsernameExists {
		response := HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	existingUser, err := m.Datastore.GetUserByCredentials(loginUser)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)

	}

	response := user.RegisterResponse{
		Uuid:     existingUser.Uuid,
		Username: cleanedUser.Username,
	}
	return c.JSON(http.StatusOK, response)
}

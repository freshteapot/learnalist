package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/api/authenticate"
	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/labstack/echo/v4"
)

type HttpRegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HttpRegisterResponse struct {
	Uuid string `json:"uuid"`
}

/*
When a user is created it returns a 201.
When a user is created with the same username and password it returns a 200.
When a user is created with a username in the system it returns a 400.
*/
func (env *Env) PostRegister(c echo.Context) error {
	var input = &HttpRegisterInput{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)
	if err != nil {
		response := HttpResponseMessage{
			Message: "Bad input.",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	if input.Username == "" || input.Password == "" {
		response := HttpResponseMessage{
			Message: "Bad input.",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	loginUser := authenticate.LoginUser{
		Username: input.Username,
		Password: input.Password,
	}

	statusCode := http.StatusCreated
	newUser, err := env.Datastore.GetUserByCredentials(loginUser)
	if err != nil {
		if err.Error() == i18n.DatabaseLookupNotFound {
			newUser, err = env.Datastore.InsertNewUser(loginUser)
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

	response := HttpRegisterResponse{
		Uuid: newUser.Uuid}
	return c.JSON(statusCode, response)
}

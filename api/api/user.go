package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/api/authenticate"
	"github.com/labstack/echo/v4"
)

type HttpRegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HttpRegisterResponse struct {
	Uuid string `json:"uuid"`
}

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
	newUser, err := env.Datastore.InsertNewUser(loginUser)
	if err != nil {
		// Currently returns Failed to save. I wonder
		// if it should give some more detail, as it
		// is actually failing as the password doesnt
		// match the username.
		response := HttpResponseMessage{
			Message: err.Error(),
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	response := HttpRegisterResponse{
		Uuid: newUser.Uuid}
	return c.JSON(http.StatusOK, response)
}

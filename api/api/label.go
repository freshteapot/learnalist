package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/freshteapot/learnalist-api/api/models"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
)

type HttpLabelInput struct {
	Label string `json:"label"`
}

func (env *Env) PostUserLabel(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	var input = &HttpLabelInput{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.PostUserLabelJSONFailure,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	label := models.NewUserLabel(input.Label, user.Uuid)
	statusCode, err := env.Datastore.PostUserLabel(label)
	if err != nil {
		if statusCode != http.StatusOK {
			response := HttpResponseMessage{
				Message: err.Error(),
			}
			return c.JSON(statusCode, response)
		}
	}
	labels, _ := env.Datastore.GetUserLabels(user.Uuid)
	return c.JSON(statusCode, labels)
}

func (env *Env) GetUserLabels(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	labels, err := env.Datastore.GetUserLabels(user.Uuid)
	if err != nil {
		// TODO log this
		response := HttpResponseMessage{
			Message: err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, labels)
}

func (env *Env) RemoveUserLabel(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	r := c.Request()
	// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
	label := strings.TrimPrefix(r.URL.Path, "/labels/")
	fmt.Println("Sad times to need to do it.")
	err := env.Datastore.RemoveUserLabel(label, user.Uuid)
	response := HttpResponseMessage{}
	if err != nil {
		response.Message = err.Error()
		return c.JSON(http.StatusInternalServerError, response)
	}
	response.Message = fmt.Sprintf(i18n.DeleteUserLabelSuccess, label)
	return c.JSON(http.StatusOK, response)
}

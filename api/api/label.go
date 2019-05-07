package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/api/models"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
)

type HttpLabelInput struct {
	Label string `json:"label"`
}

func (env *Env) GetLabelsByUser(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	labels := env.Datastore.GetLabelsByUser(user.Uuid)
	return c.JSON(http.StatusOK, labels)
}

func (env *Env) PostLabel(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	var input = &HttpLabelInput{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)

	if err != nil {
		response := HttpResponseMessage{
			Message: "Bad input.",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	// Does this user already have this label?
	labels := env.Datastore.GetLabelsByUser(user.Uuid)
	for _, label := range labels {
		if label.Label == input.Label {
			return c.JSON(http.StatusOK, label)
		}
	}

	label := models.NewLabel()
	label.Label = input.Label
	label.UserUuid = user.Uuid

	err = env.Datastore.SaveLabel(*label)
	if err != nil {
		response := HttpResponseMessage{
			Message: err.Error(),
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	return c.JSON(http.StatusCreated, label)
}

func (env *Env) RemoveLabel(c echo.Context) error {
	var message string
	var err error
	var label *models.Label
	response := HttpResponseMessage{}

	labelUuid := c.Param("uuid")
	r := c.Request()
	hackUuid := strings.TrimPrefix(r.URL.Path, "/labels/")
	fmt.Println("Sad times to need to do it.")
	fmt.Println(labelUuid)
	fmt.Println(hackUuid)
	labelUuid = hackUuid

	// Check to see if user has access
	user := c.Get("loggedInUser").(uuid.User)
	label, err = env.Datastore.GetLabel(labelUuid)

	if err != nil {
		return c.NoContent(http.StatusNoContent)
	}

	if label.UserUuid != user.Uuid {
		response.Message = "You are not the owner of this label."
		return c.JSON(http.StatusForbidden, response)
	}

	err = env.Datastore.RemoveLabel(labelUuid)

	message = fmt.Sprintf("Label %s was removed.", labelUuid)
	if err != nil {
		message = fmt.Sprintf("Failed to remove label with error %s", err.Error())
		response.Message = message
		return c.JSON(http.StatusInternalServerError, response)
	}
	response.Message = message
	return c.JSON(http.StatusOK, response)
}

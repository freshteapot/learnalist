package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1PostUserLabel(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	var input = &api.HttpLabelInput{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)
	if err != nil {
		response := api.HttpResponseMessage{
			Message: i18n.PostUserLabelJSONFailure,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	label := label.NewUserLabel(input.Label, user.Uuid)
	statusCode, err := m.Datastore.Labels().PostUserLabel(label)
	switch statusCode {
	case http.StatusOK:
		break
	case http.StatusCreated:
		break
	case http.StatusBadRequest:
		response := api.HttpResponseMessage{
			Message: i18n.ValidationLabel,
		}
		return c.JSON(http.StatusBadRequest, response)
	case http.StatusInternalServerError:
		fallthrough
	default:
		response := api.HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	labels, _ := m.Datastore.Labels().GetUserLabels(user.Uuid)
	return c.JSON(statusCode, labels)
}

func (m *Manager) V1GetUserLabels(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	labels, err := m.Datastore.Labels().GetUserLabels(user.Uuid)
	if err != nil {
		// TODO log this
		response := api.HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	return c.JSON(http.StatusOK, labels)
}

func (m *Manager) V1RemoveUserLabel(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	label := c.Param("label")

	err := m.Datastore.RemoveUserLabel(label, user.Uuid)
	response := api.HttpResponseMessage{}
	if err != nil {
		// TODO log this
		response := api.HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	response.Message = fmt.Sprintf(i18n.ApiDeleteUserLabelSuccess, label)
	return c.JSON(http.StatusOK, response)
}

package api

import (
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1SaveAlist(c echo.Context) error {
	var inputUuid string
	user := c.Get("loggedInUser").(uuid.User)
	method := c.Request().Method

	switch method {
	case http.MethodPost:
		break
	case http.MethodPut:
		break
	default:
		response := api.HttpResponseMessage{
			Message: i18n.ApiMethodNotSupported,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	var aList alist.Alist
	err := aList.UnmarshalJSON(jsonBytes)
	if err != nil {
		response := api.HttpResponseMessage{
			Message: i18n.InputAlistJSONFailure,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	aList.User = user
	if method == http.MethodPut {
		inputUuid = c.Param("uuid")
		if inputUuid == "" {
			response := api.HttpResponseMessage{
				Message: i18n.ValidationAlists,
			}
			return c.JSON(http.StatusBadRequest, response)
		}

		if aList.Uuid != "" {
			if inputUuid != aList.Uuid {
				response := api.HttpResponseMessage{
					Message: i18n.ValidationUUIDMismatch,
				}
				return c.JSON(http.StatusBadRequest, response)
			}
		}
		aList.Uuid = inputUuid
	}

	aList, err = m.Datastore.SaveAlist(method, aList)
	if err != nil {
		response := api.HttpResponseMessage{
			Message: err.Error(),
		}

		switch err {
		case i18n.ErrorListNotFound:
			return c.JSON(http.StatusNotFound, response)
		case i18n.ErrorAListFromDomainMisMatch:
			fallthrough
		case i18n.ErrorInputSaveAlistOperationFromRestriction:
			fallthrough
		case i18n.ErrorInputSaveAlistFromKindNotSupported:
			return c.JSON(http.StatusUnprocessableEntity, response)
		}

		// This is a little ugly
		switch err.Error() {
		case i18n.InputSaveAlistOperationOwnerOnly: // Maybe move this to 422
			return c.JSON(http.StatusForbidden, response)
		default:
			return c.JSON(http.StatusBadRequest, response)
		}
	}

	// Save to hugo
	m.HugoHelper.WriteList(aList)
	// TODO this might become a painful bottle neck
	m.HugoHelper.WriteListsByUser(aList.User.Uuid, m.Datastore.GetAllListsByUser(user.Uuid))
	m.HugoHelper.WritePublicLists(m.Datastore.GetPublicLists())

	statusCode := http.StatusOK
	if method == http.MethodPost {
		statusCode = http.StatusCreated
	}
	return c.JSON(statusCode, aList)
}

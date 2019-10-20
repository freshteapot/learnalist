package api

import (
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
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
		response := HttpResponseMessage{
			Message: i18n.ApiMethodNotSupported,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	aList := new(alist.Alist)
	err := aList.UnmarshalJSON(jsonBytes)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.InputAlistJSONFailure,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	aList.User = user

	if method == http.MethodPut {
		inputUuid = c.Param("uuid")
		if inputUuid == "" {
			response := HttpResponseMessage{
				Message: i18n.ValidationAlists,
			}
			return c.JSON(http.StatusBadRequest, response)
		}

		if aList.Uuid != "" {
			if inputUuid != aList.Uuid {
				response := HttpResponseMessage{
					Message: i18n.ValidationUUIDMismatch,
				}
				return c.JSON(http.StatusBadRequest, response)
			}
		}
		aList.Uuid = inputUuid
	}

	aList, err = m.Datastore.SaveAlist(method, *aList)
	if err != nil {
		switch err.Error() {
		case i18n.SuccessAlistNotFound:
			response := HttpResponseMessage{
				Message: i18n.SuccessAlistNotFound,
			}
			return c.JSON(http.StatusNotFound, response)
		case i18n.InputSaveAlistOperationOwnerOnly:
			response := HttpResponseMessage{
				Message: i18n.InputSaveAlistOperationOwnerOnly,
			}
			return c.JSON(http.StatusForbidden, response)
		default:
			response := HttpResponseMessage{
				Message: err.Error(),
			}
			return c.JSON(http.StatusBadRequest, response)
		}
	}

	// Save to hugo
	m.HugoHelper.Write(aList)
	statusCode := http.StatusOK
	if method == http.MethodPost {
		statusCode = http.StatusCreated
	}
	return c.JSON(statusCode, *aList)
}

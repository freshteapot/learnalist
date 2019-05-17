package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/api/alist"
	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
)

func (env *Env) GetListsByMe(c echo.Context) error {
	var alists []*alist.Alist
	user := c.Get("loggedInUser").(uuid.User)
	r := c.Request()
	params := r.URL.Query()
	filterByLabels := params.Get("labels")
	listType := params.Get("list_type")
	alists = env.Datastore.GetListsByUserWithFilters(user.Uuid, filterByLabels, listType)
	return c.JSON(http.StatusOK, alists)
}

func (env *Env) GetListByUUID(c echo.Context) error {
	r := c.Request()
	// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
	uuid := strings.TrimPrefix(r.URL.Path, "/alist/")
	if uuid == "" {
		response := HttpResponseMessage{
			Message: i18n.InputMissingListUuid,
		}
		return c.JSON(http.StatusNotFound, response)
	}
	alist, err := env.Datastore.GetAlist(uuid)
	if err != nil {
		message := fmt.Sprintf(i18n.ApiAlistNotFound, uuid)
		response := HttpResponseMessage{
			Message: message,
		}
		return c.JSON(http.StatusNotFound, response)
	}
	return c.JSON(http.StatusOK, *alist)
}

func (env *Env) SaveAlist(c echo.Context) error {
	var inputUuid string
	user := c.Get("loggedInUser").(uuid.User)
	method := c.Request().Method

	if method != http.MethodPost {
		if method != http.MethodPut {
			response := HttpResponseMessage{
				Message: i18n.ApiMethodNotSupported,
			}
			return c.JSON(http.StatusBadRequest, response)
		}
	}

	if method == http.MethodPut {
		// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
		r := c.Request()
		inputUuid = strings.TrimPrefix(r.URL.Path, "/alist/")
	}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	aList := new(alist.Alist)
	err := aList.UnmarshalJSON(jsonBytes)
	if err != nil {
		message := fmt.Sprintf("Your Json has a problem. %s", err)
		response := HttpResponseMessage{
			Message: message,
		}

		return c.JSON(http.StatusBadRequest, response)
	}

	aList.User = user

	if method == http.MethodPut {
		aList.Uuid = inputUuid
	}

	aList, err = env.Datastore.SaveAlist(method, *aList)
	if err != nil {
		response := HttpResponseMessage{
			Message: err.Error(),
		}

		if err.Error() == i18n.InputSaveAlistOperationOwnerOnly {
			return c.JSON(http.StatusForbidden, response)
		}

		return c.JSON(http.StatusBadRequest, response)
	}

	statusCode := http.StatusOK
	if method == http.MethodPost {
		statusCode = http.StatusCreated
	}
	return c.JSON(statusCode, *aList)
}

func (env *Env) RemoveAlist(c echo.Context) error {
	r := c.Request()
	// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
	alist_uuid := strings.TrimPrefix(r.URL.Path, "/alist/")

	user := c.Get("loggedInUser").(uuid.User)
	response := HttpResponseMessage{}

	err := env.Datastore.RemoveAlist(alist_uuid, user.Uuid)
	if err != nil {
		if err.Error() == i18n.SuccessAlistNotFound {
			response.Message = err.Error()
			return c.JSON(http.StatusNotFound, response)
		}

		if err.Error() == i18n.InputDeleteAlistOperationOwnerOnly {
			response.Message = err.Error()
			return c.JSON(http.StatusForbidden, response)
		}

		response.Message = i18n.InternalServerErrorDeleteAlist
		return c.JSON(http.StatusInternalServerError, response)
	}

	response.Message = fmt.Sprintf(i18n.ApiDeleteAlistSuccess, alist_uuid)
	return c.JSON(http.StatusOK, response)
}

package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

/*
@Param labels filter lists by label: "car", "car,bil".
@Param list_type filter lists by type: "v1", "v2".
*/
func (m *Manager) V1GetListsByMe(c echo.Context) error {
	var alists []*alist.Alist
	user := c.Get("loggedInUser").(uuid.User)
	r := c.Request()
	params := r.URL.Query()
	filterByLabels := params.Get("labels")
	listType := params.Get("list_type")
	alists = m.Datastore.GetListsByUserWithFilters(user.Uuid, filterByLabels, listType)
	return c.JSON(http.StatusOK, alists)
}

func (m *Manager) V1GetListByUUID(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	r := c.Request()
	// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
	uuid := strings.TrimPrefix(r.URL.Path, "/api/v1/alist/")

	if uuid == "" {
		response := HttpResponseMessage{
			Message: i18n.InputMissingListUuid,
		}
		return c.JSON(http.StatusNotFound, response)
	}

	allow, err := m.Acl.HasUserListReadAccess(uuid, user.Uuid)
	if err != nil {
		response := HttpResponseMessage{
			Message: "I broke something",
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	if !allow {
		response := HttpResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	alist, err := m.Datastore.GetAlist(uuid)
	if err != nil {
		message := fmt.Sprintf(i18n.ApiAlistNotFound, uuid)
		response := HttpResponseMessage{
			Message: message,
		}
		return c.JSON(http.StatusNotFound, response)
	}

	return c.JSON(http.StatusOK, *alist)
}

func (m *Manager) V1SaveAlist(c echo.Context) error {
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
		inputUuid = strings.TrimPrefix(r.URL.Path, "/api/v1/alist/")
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

	aList, err = m.Datastore.SaveAlist(method, *aList)
	if err != nil {
		response := HttpResponseMessage{
			Message: err.Error(),
		}

		if err.Error() == i18n.InputSaveAlistOperationOwnerOnly {
			return c.JSON(http.StatusForbidden, response)
		}

		return c.JSON(http.StatusBadRequest, response)
	}

	// Save to hugo
	m.HugoHelper.Write(aList)
	statusCode := http.StatusOK
	if method == http.MethodPost {
		statusCode = http.StatusCreated
	}
	return c.JSON(statusCode, *aList)
}

func (m *Manager) V1RemoveAlist(c echo.Context) error {
	r := c.Request()
	// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
	alist_uuid := strings.TrimPrefix(r.URL.Path, "/api/v1/alist/")

	user := c.Get("loggedInUser").(uuid.User)
	response := HttpResponseMessage{}

	err := m.Datastore.RemoveAlist(alist_uuid, user.Uuid)
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

	// Remove from cache
	m.HugoHelper.Remove(alist_uuid)
	response.Message = fmt.Sprintf(i18n.ApiDeleteAlistSuccess, alist_uuid)
	return c.JSON(http.StatusOK, response)
}

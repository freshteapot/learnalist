package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1ShareListReadAccess(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	// TODO maybe we support an array
	var input = &HttpShareListWithUserInput{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.PostShareListJSONFailure,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	switch input.Action {
	case aclKeys.ActionGrant:
		break
	case aclKeys.ActionRevoke:
		break
	default:
		response := HttpResponseMessage{
			Message: i18n.ApiShareValidationError,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	aList, err := m.Datastore.GetAlist(input.AlistUUID)
	if err != nil {
		if err.Error() == i18n.SuccessAlistNotFound {
			response := HttpResponseMessage{
				Message: i18n.SuccessAlistNotFound,
			}
			return c.JSON(http.StatusNotFound, response)
		}

		response := HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	if aList.Info.SharedWith == aclKeys.NotShared {
		return c.JSON(http.StatusBadRequest, HttpResponseMessage{
			Message: i18n.ApiShareReadAccessInvalidWithNotShared,
		})
	}

	if aList.User.Uuid != user.Uuid {
		response := HttpResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	if input.UserUUID == user.Uuid {
		response := HttpResponseMessage{
			Message: i18n.ApiShareYouCantShareWithYourself,
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	if !m.Datastore.UserExists(input.UserUUID) {
		response := HttpResponseMessage{
			Message: i18n.SuccessUserNotFound,
		}
		return c.JSON(http.StatusNotFound, response)
	}

	switch input.Action {
	case aclKeys.ActionGrant:
		m.Acl.GrantUserListReadAccess(input.AlistUUID, input.UserUUID)
	case aclKeys.ActionRevoke:
		m.Acl.RevokeUserListReadAccess(input.AlistUUID, input.UserUUID)
	}

	return c.JSON(http.StatusOK, input)
}

func (m *Manager) V1ShareAlist(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	var input = &HttpShareListInput{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.PostShareListJSONFailure,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	switch input.Action {
	case aclKeys.SharedWithPublic:
		break
	case aclKeys.NotShared:
		break
	case aclKeys.SharedWithFriends:
		break
	default:
		response := HttpResponseMessage{
			Message: i18n.ApiShareValidationError,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	aList, _ := m.Datastore.GetAlist(input.AlistUUID)
	if aList.Uuid == "" {
		response := HttpResponseMessage{
			Message: i18n.SuccessAlistNotFound,
		}
		return c.JSON(http.StatusNotFound, response)
	}

	if aList.User.Uuid != user.Uuid {
		response := HttpResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	if aList.Info.SharedWith == input.Action {
		return c.JSON(http.StatusOK, HttpResponseMessage{
			Message: i18n.ApiShareNoChange,
		})
	}

	aList.Info.SharedWith = input.Action
	m.Datastore.SaveAlist(http.MethodPut, aList)
	// Save to hugo
	m.HugoHelper.WriteList(aList)
	aLists := m.Datastore.GetListsByUserWithFilters(aList.User.Uuid, "", "")
	m.HugoHelper.WriteListsByUser(aList.User.Uuid, aLists)

	message := ""
	switch input.Action {
	case aclKeys.SharedWithPublic:
		message = i18n.ApiShareListSuccessWithPublic
	case aclKeys.NotShared:
		message = i18n.ApiShareListSuccessPrivate
	case aclKeys.SharedWithFriends:
		message = i18n.ApiShareListSuccessWithFriends
	}

	return c.JSON(http.StatusOK, HttpResponseMessage{
		Message: message,
	})
}

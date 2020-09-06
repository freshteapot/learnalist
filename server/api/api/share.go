package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
)

// TODO CHECK this function for "from too"
func (m *Manager) V1ShareListReadAccess(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	// TODO maybe we support an array
	var input = &api.HttpShareListWithUserInput{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)
	if err != nil {
		response := api.HttpResponseMessage{
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
		response := api.HttpResponseMessage{
			Message: i18n.ApiShareValidationError,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	aList, err := m.Datastore.GetAlist(input.AlistUUID)
	if err != nil {
		if err.Error() == i18n.SuccessAlistNotFound {
			response := api.HttpResponseMessage{
				Message: i18n.SuccessAlistNotFound,
			}
			return c.JSON(http.StatusNotFound, response)
		}

		response := api.HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	if aList.User.Uuid != user.Uuid {
		response := api.HttpResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	if !alist.WithFromCheckSharing(aList.Info) {
		response := api.HttpResponseMessage{
			Message: i18n.InputSaveAlistOperationFromRestriction,
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	if aList.Info.SharedWith == aclKeys.NotShared {
		return c.JSON(http.StatusBadRequest, api.HttpResponseMessage{
			Message: i18n.ApiShareReadAccessInvalidWithNotShared,
		})
	}

	if input.UserUUID == user.Uuid {
		response := api.HttpResponseMessage{
			Message: i18n.ApiShareYouCantShareWithYourself,
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	if !m.Datastore.UserExists(input.UserUUID) {
		response := api.HttpResponseMessage{
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

// TODO CHECK this function for "from too"
func (m *Manager) V1ShareAlist(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	var input = &api.HttpShareListInput{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)
	if err != nil {
		response := api.HttpResponseMessage{
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
		response := api.HttpResponseMessage{
			Message: i18n.ApiShareValidationError,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	aList, _ := m.Datastore.GetAlist(input.AlistUUID)
	if aList.Uuid == "" {
		response := api.HttpResponseMessage{
			Message: i18n.SuccessAlistNotFound,
		}
		return c.JSON(http.StatusNotFound, response)
	}

	if aList.User.Uuid != user.Uuid {
		response := api.HttpResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	// start: Check FromSharing
	checkInfo := aList.Info
	checkInfo.SharedWith = input.Action
	if !alist.WithFromCheckSharing(checkInfo) {
		response := api.HttpResponseMessage{
			Message: i18n.InputSaveAlistOperationFromRestriction,
		}
		return c.JSON(http.StatusForbidden, response)
	}
	// end: Check FromSharing

	if aList.Info.SharedWith == input.Action {
		return c.JSON(http.StatusOK, api.HttpResponseMessage{
			Message: i18n.ApiShareNoChange,
		})
	}

	// Checks passed, now we update the value
	aList.Info.SharedWith = input.Action
	m.Datastore.SaveAlist(http.MethodPut, aList)
	// Save to hugo
	m.HugoHelper.WriteList(aList)
	// TODO this might become a painful bottle neck
	m.HugoHelper.WriteListsByUser(aList.User.Uuid, m.Datastore.GetAllListsByUser(user.Uuid))
	m.HugoHelper.WritePublicLists(m.Datastore.GetPublicLists())

	message := ""
	switch input.Action {
	case aclKeys.SharedWithPublic:
		message = i18n.ApiShareListSuccessWithPublic
	case aclKeys.NotShared:
		message = i18n.ApiShareListSuccessPrivate
	case aclKeys.SharedWithFriends:
		message = i18n.ApiShareListSuccessWithFriends
	}

	return c.JSON(http.StatusOK, api.HttpResponseMessage{
		Message: message,
	})
}

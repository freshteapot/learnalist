package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

const (
	ActionRevoke           = "revoke"
	ActionGrant            = "grant"
	ActionShareWithPublic  = "public"
	ActionShareWithOwner   = "private"
	ActionShareWithPrivate = "friends"
)

type HttpShareListReadAccessInput struct {
	AlistUUID string `json:"alist_uuid"`
	Action    string `json:"action"`
}

type HttpShareListWithUserInput struct {
	UserUUID  string `json:"user_uuid"`
	AlistUUID string `json:"alist_uuid"`
	Action    string `json:"action"`
}

func (m *Manager) V1ShareAlist(c echo.Context) error {
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

	aList, _ := m.Datastore.GetAlist(input.AlistUUID)
	if aList == nil {
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

	if input.UserUUID != user.Uuid {
		if !m.Datastore.UserExists(input.UserUUID) {
			response := HttpResponseMessage{
				Message: i18n.SuccessUserNotFound,
			}
			return c.JSON(http.StatusNotFound, response)
		}
		if input.Action == ActionGrant {
			m.Acl2.GrantUserListReadAccess(input.AlistUUID, input.UserUUID)
		}
		if input.Action == ActionRevoke {
			m.Acl2.RevokeUserListReadAccess(input.AlistUUID, input.UserUUID)
		}
	}

	return c.JSON(http.StatusOK, input)
}

func (m *Manager) V1ShareListReadAccess(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	var input = &HttpShareListReadAccessInput{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, input)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.PostShareListJSONFailure,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	aList, _ := m.Datastore.GetAlist(input.AlistUUID)
	if aList == nil {
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

	message := ""
	if input.Action == ActionShareWithPublic {
		m.Acl2.ShareListWithPublic(aList.Uuid)
		message = "List is now public"
	}

	if input.Action == ActionShareWithOwner {
		m.Acl2.MakeListPrivate(aList.Uuid, aList.User.Uuid)
		message = "List is now private to the owner"
	}

	if input.Action == ActionShareWithPrivate {
		m.Acl2.ShareListWithFriends(aList.Uuid)
		message = "List is now private to the owner and those granted access"
	}

	response := HttpResponseMessage{
		Message: message,
	}
	return c.JSON(http.StatusOK, response)
}

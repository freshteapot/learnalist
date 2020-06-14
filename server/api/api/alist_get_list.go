package api

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
)

/*
@Param labels filter lists by label: "car", "car,bil".
@Param list_type filter lists by type: "v1", "v2".
*/
func (m *Manager) V1GetListsByMe(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	r := c.Request()
	params := r.URL.Query()
	filterByLabels := params.Get("labels")
	listType := params.Get("list_type")
	aLists := m.Datastore.GetListsByUserWithFilters(user.Uuid, filterByLabels, listType)
	return c.JSON(http.StatusOK, aLists)
}

func (m *Manager) V1GetListByUUID(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	uuid := c.Param("uuid")
	if uuid == "" {
		response := api.HttpResponseMessage{
			Message: i18n.InputMissingListUuid,
		}
		return c.JSON(http.StatusNotFound, response)
	}

	allow, err := m.Acl.HasUserListReadAccess(uuid, user.Uuid)
	if err != nil {
		response := api.HttpResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	if !allow {
		response := api.HttpResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	alist, err := m.Datastore.GetAlist(uuid)
	if err != nil {
		// With the look up of access before, this one doesnt fully make sense anymore.
		message := fmt.Sprintf(i18n.ApiAlistNotFound, uuid)
		response := api.HttpResponseMessage{
			Message: message,
		}
		return c.JSON(http.StatusNotFound, response)
	}

	return c.JSON(http.StatusOK, alist)
}

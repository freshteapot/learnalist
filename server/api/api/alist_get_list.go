package api

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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
		response := api.HTTPResponseMessage{
			Message: i18n.InputMissingListUuid,
		}
		return c.JSON(http.StatusNotFound, response)
	}

	allow, err := m.Acl.HasUserListReadAccess(uuid, user.Uuid)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}
	if !allow {
		response := api.HTTPResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	alist, err := m.Datastore.GetAlist(uuid)
	if err != nil {
		if err == i18n.ErrorListNotFound {
			m.logger.WithFields(logrus.Fields{
				"event":      "broken-state",
				"alist_uuid": uuid,
			}).Error("List not found, but has acl access")

			message := fmt.Sprintf(i18n.ApiAlistNotFound, uuid)
			response := api.HTTPResponseMessage{
				Message: message,
			}
			return c.JSON(http.StatusNotFound, response)
		}
		// When the db fails to lookup, maybe we should actually be crashing.
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	return c.JSON(http.StatusOK, alist)
}

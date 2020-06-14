package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

func (m *Manager) GetAlist(c echo.Context) error {
	var err error
	var alistUUID string
	var isA string
	publicFolder := m.HugoHelper.GetPubicDirectory()

	uri := c.Request().URL.Path
	user := c.Get("loggedInUser")
	userUUID := ""
	if user != nil {
		userUUID = user.(uuid.User).Uuid
	}
	alistUUID, isA, err = GetAlistUUIDFromURL(uri)

	if err != nil {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alist/404.html", publicFolder))
		return c.HTMLBlob(http.StatusNotFound, data)
	}

	// TODO should we check if it exists?
	allow, err := m.Acl.HasUserListReadAccess(alistUUID, userUUID)
	if err != nil {
		// TODO log this?
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alist/500.html", publicFolder))
		return c.HTMLBlob(http.StatusInternalServerError, data)
	}

	if !allow {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alist/no-access.html", publicFolder))
		return c.HTMLBlob(http.StatusForbidden, data)
	}

	// At this point, we assume the list is real
	pathToAlist := fmt.Sprintf("%s/alist/%s.%s", publicFolder, alistUUID, isA)
	if _, err := os.Stat(pathToAlist); err == nil {
		return c.File(pathToAlist)
	}

	return c.File(fmt.Sprintf("%s/alist/please-refresh.html", publicFolder))
}

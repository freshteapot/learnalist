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

	uri := c.Request().URL.Path
	user := c.Get("loggedInUser")
	userUUID := ""
	if user != nil {
		userUUID = user.(uuid.User).Uuid
	}

	alistUUID, isA, err = GetAlistUUIDFromUrl(uri)

	if err != nil {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alists/404.html", m.SiteCacheFolder))
		return c.HTMLBlob(http.StatusNotFound, data)
	}

	// TODO should we check if it exists?
	allow, err := m.Acl.HasUserListReadAccess(alistUUID, userUUID)
	if err != nil {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alists/500.html", m.SiteCacheFolder))
		return c.HTMLBlob(http.StatusInternalServerError, data)
	}

	if !allow {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alists/no-access.html", m.SiteCacheFolder))
		return c.HTMLBlob(http.StatusForbidden, data)
	}

	// At this point, we assume the list is real
	// This code should only serve the lists?
	pathToAlist := fmt.Sprintf("%s/alists/%s.%s", m.SiteCacheFolder, alistUUID, isA)

	if _, err := os.Stat(pathToAlist); err == nil {
		return c.File(pathToAlist)
	}

	// Assumed it is a fast request (maybe e2e tests) so blindly trigger a new Build
	// of any queued content.
	// This is a really ugly hack :P
	// The e2e client bubbled this up as I have the timeouts set stupidly low
	go m.HugoHelper.ProcessContent()

	data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alists/please-refresh.html", m.SiteCacheFolder))
	return c.HTMLBlob(http.StatusOK, data)

	// TODO handle html or json
	// Maybe use HTTPErrorHandler
	// https://echo.labstack.com/guide/error-handling#custom-http-error-handler
	// pathToFile = fmt.Sprintf("%s/404.html", m.SiteCacheFolder)
	// return c.File(pathToFile)
}

package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

func (m *Manager) GetMyLists(c echo.Context) error {
	// TODO is this good enough

	user := c.Get("loggedInUser")
	if user == nil {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alist/no-access.html", m.SiteCacheFolder))
		return c.HTMLBlob(http.StatusForbidden, data)
	}
	userUUID := user.(uuid.User).Uuid

	// At this point, we assume the user is real
	pathToAlist := fmt.Sprintf("%s/alistsbyuser/%s.html", m.SiteCacheFolder, userUUID)

	_, err := os.Stat(pathToAlist)

	if err != nil {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alist/500.html", m.SiteCacheFolder))
		return c.HTMLBlob(http.StatusInternalServerError, data)

	}

	return c.File(pathToAlist)
}

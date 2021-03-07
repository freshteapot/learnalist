package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

func (m *Manager) GetMyLists(c echo.Context) error {
	publicFolder := m.pathToPublicDirectory
	user := c.Get("loggedInUser")
	if user == nil {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alist/no-access.html", publicFolder))
		return c.HTMLBlob(http.StatusForbidden, data)
	}
	userUUID := user.(uuid.User).Uuid

	// At this point, we assume the user is real
	pathToAlist := fmt.Sprintf("%s/alistsbyuser/%s.html", publicFolder, userUUID)

	_, err := os.Stat(pathToAlist)

	if err != nil {
		// TODO something is broken, if this happens
		// How do we handle first time users?
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alist/500.html", publicFolder))
		return c.HTMLBlob(http.StatusInternalServerError, data)
	}

	return c.File(pathToAlist)
}

func (m *Manager) GetMyListsByURI(c echo.Context) error {
	publicFolder := m.pathToPublicDirectory
	// If the userID in the url matches the one via loggedInUser
	// then let them see the list, if not reject.
	url := c.Request().URL.Path
	url = strings.TrimPrefix(url, "/alistsbyuser/")
	userUUIDviaURL := strings.TrimSuffix(url, ".html")

	user := c.Get("loggedInUser")
	if user == nil {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alist/no-access.html", publicFolder))
		return c.HTMLBlob(http.StatusForbidden, data)
	}
	userUUID := user.(uuid.User).Uuid

	if userUUID != userUUIDviaURL {
		data, _ := ioutil.ReadFile(fmt.Sprintf("%s/alist/no-access.html", publicFolder))
		return c.HTMLBlob(http.StatusForbidden, data)
	}

	return m.GetMyLists(c)
}

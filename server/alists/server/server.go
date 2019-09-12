package server

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

type HttpResponseMessage struct {
	Message string `json:"message"`
}

type Manager struct {
	Acl             acl.Acl
	Datastore       models.Datastore
	SiteCacheFolder string
	HugoHelper      hugo.HugoHelper
}

type ErrorHttpCode int

func (m *Manager) GetPlay(c echo.Context) error {
	// TODO which list,
	// TODO which list type,
	// TODO what to include.
	// TODO can this be static rendered based on the object?
	response := `
function setup() {
	// Turn the list into an object
	let aList = JSON.parse(JSON.parse(document.querySelector("#data").innerHTML));

	document.querySelector("#play").style.display = "none";
	var interact = document.createElement("v1-slideshow");
	interact.aList = aList;
	document.querySelector("#play").appendChild(interact);
}
setup()
`
	c.Response().Header().Set("Content-Type", "application/javascript")
	return c.String(http.StatusOK, response)

}

func (m *Manager) GetAlist(c echo.Context) error {
	var pathToFile string
	var err ErrorHttpCode

	uri := c.Request().URL.Path
	user := c.Get("loggedInUser")
	userUUID := ""
	if user != nil {
		userUUID = user.(uuid.User).Uuid
	}
	pathToFile, err = m.serveAlist(userUUID, uri)
	if pathToFile != "" {
		return c.File(pathToFile)
	}

	if err == http.StatusForbidden {
		//TODO use a better warning
		return c.String(http.StatusForbidden, "Not allowed access")
	}

	pathToFile, _ = m.serveStatic(uri)
	if pathToFile != "" {
		return c.File(pathToFile)
	}

	// TODO handle html or json
	// Maybe use HTTPErrorHandler
	// https://echo.labstack.com/guide/error-handling#custom-http-error-handler
	pathToFile = fmt.Sprintf("%s/404.html", m.SiteCacheFolder)
	return c.File(pathToFile)
}

func (m *Manager) serveAlist(userUUID string, urlPath string) (string, ErrorHttpCode) {
	parts := strings.Split(urlPath, "/")
	suffix := parts[len(parts)-1]
	parts = strings.Split(suffix, ".")
	if len(parts) != 2 {
		return "", http.StatusFound
	}

	alistUUID := parts[0]
	isA := parts[1]
	// This code should only serve the lists?
	path := fmt.Sprintf("%s/alists/%s.%s", m.SiteCacheFolder, alistUUID, isA)

	if _, err := os.Stat(path); err == nil {
		if !m.Acl.HasUserListReadAccess(userUUID, alistUUID) {
			return "", http.StatusForbidden
		}
		return path, http.StatusOK
	}
	return "", http.StatusNotFound
}

func (m *Manager) serveStatic(urlPath string) (string, ErrorHttpCode) {
	path := fmt.Sprintf("%s/%s", m.SiteCacheFolder, urlPath[1:])
	if _, err := os.Stat(path); err == nil {
		return path, http.StatusOK
	}
	return "", http.StatusNotFound
}

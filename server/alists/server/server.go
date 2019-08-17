package server

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/labstack/echo/v4"
)

type HttpResponseMessage struct {
	Message string `json:"message"`
}

type Manager struct {
	Acl              acl.Acl
	StaticSiteFolder string
}

func (m *Manager) GetAlist(c echo.Context) error {
	var pathToFile string
	var err error
	pathToFile, err = m.serveAlist(c.Request().URL.Path)
	if pathToFile != "" {
		return c.File(pathToFile)
	}

	pathToFile, err = m.serveStatic(c.Request().URL.Path)
	if err == nil {
		return c.File(pathToFile)
	}
	// TODO handle html or json
	// TODO maybe better than nocontent
	return c.NoContent(http.StatusNotFound)
}

func (m *Manager) serveAlist(urlPath string) (string, error) {
	parts := strings.Split(urlPath, "/")
	suffix := parts[len(parts)-1]
	parts = strings.Split(suffix, ".")
	if len(parts) != 2 {
		return "", errors.New("List not found")
	}

	uuid := parts[0]
	isA := parts[1]
	// This code should only serve the lists?
	path := fmt.Sprintf("%s/alists/%s.%s", m.StaticSiteFolder, uuid, isA)

	if _, err := os.Stat(path); err == nil {
		// TODO at this point we can do acl look up.

		// http.ServeFile(w, r, path)
		return path, nil
	}
	return "", errors.New("List not found")
}

func (m *Manager) serveStatic(urlPath string) (string, error) {
	path := fmt.Sprintf("%s/%s", m.StaticSiteFolder, urlPath[1:])
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", errors.New("File not found")
}
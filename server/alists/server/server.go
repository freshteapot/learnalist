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

type Manager struct {
	Acl              acl.Acl
	StaticSiteFolder string
}

func (m *Manager) GetAlist(c echo.Context) error {
	return m.serveFiles(c.Response().Writer, c.Request())
}

func (m *Manager) serveFiles(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil
	}

	err := m.serveAlist(w, r)
	if err == nil {
		return nil
	}
	fmt.Println("Error after serveAlist")
	fmt.Println(err)
	err = m.serveStatic(w, r)
	if err == nil {
		return nil
	}

	fmt.Println("Error after serveStatic")
	fmt.Println(err)

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "custom 404")
	return nil
}

func (m *Manager) serveAlist(w http.ResponseWriter, r *http.Request) error {
	parts := strings.Split(r.URL.Path, "/")
	suffix := parts[len(parts)-1]
	parts = strings.Split(suffix, ".")
	if len(parts) != 2 {
		return errors.New("List not found")
	}

	uuid := parts[0]
	isA := parts[1]
	// This code should only serve the lists?
	path := fmt.Sprintf("%s/alists/%s.%s", m.StaticSiteFolder, uuid, isA)

	if _, err := os.Stat(path); err == nil {
		// TODO at this point we can do acl look up.
		http.ServeFile(w, r, path)
		return nil
	}
	return errors.New("List not found")
}

func (m *Manager) serveStatic(w http.ResponseWriter, r *http.Request) error {
	// path/to/whatever does *not* exist
	path := fmt.Sprintf("%s/%s", m.StaticSiteFolder, r.URL.Path[1:])
	if _, err := os.Stat(path); err == nil {
		// path/to/whatever exists
		http.ServeFile(w, r, path)
		return nil
	}
	return errors.New("File not found")
}

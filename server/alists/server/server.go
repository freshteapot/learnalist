package server

import (
	"errors"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/models"

	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
)

type Manager struct {
	Acl                   acl.Acl
	Datastore             models.Datastore
	UserSession           user.Session
	pathToPublicDirectory string
}

func NewManager(
	acl acl.Acl,
	datastore models.Datastore,
	userSession user.Session,
	pathToPublicDirectory string,
) *Manager {
	return &Manager{
		Acl:                   acl,
		Datastore:             datastore,
		UserSession:           userSession,
		pathToPublicDirectory: pathToPublicDirectory,
	}
}

func GetAlistUUIDFromURL(input string) (string, string, error) {
	input = strings.TrimPrefix(input, "/alist/")
	if strings.Contains(input, "/") {
		return "", "", errors.New("Invalid uri")
	}

	parts := strings.Split(input, ".")
	if len(parts) != 2 {
		return "", "", errors.New("missing suffix")
	}
	alistUUID := parts[0]
	isA := parts[1]

	switch isA {
	case "html":
	case "json":
	default:
		return "", "", errors.New("Unsupported format")
	}
	return alistUUID, isA, nil
}

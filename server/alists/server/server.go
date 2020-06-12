package server

import (
	"errors"
	"strings"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/models"

	"github.com/freshteapot/learnalist-api/server/pkg/acl"
)

type HttpResponseMessage struct {
	Message string `json:"message"`
}

type Manager struct {
	Acl        acl.Acl
	Datastore  models.Datastore
	HugoHelper hugo.HugoHelper
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

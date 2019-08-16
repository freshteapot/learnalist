package api

import (
	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/freshteapot/learnalist-api/server/api/models"
)

// m exposing the data abstraction layer
type Manager struct {
	Datastore    models.Datastore
	Acl          acl.Acl
	DatabaseName string
}

type HttpResponseMessage struct {
	Message string `json:"message"`
}

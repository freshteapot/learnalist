package api

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/freshteapot/learnalist-api/server/api/models"
)

// m exposing the data abstraction layer
type Manager struct {
	Datastore    models.Datastore
	Acl          acl.Acl
	DatabaseName string
	HugoHelper   hugo.HugoSiteBuilder
}

type HttpResponseMessage struct {
	Message string `json:"message"`
}

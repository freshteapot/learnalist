package api

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/freshteapot/learnalist-api/server/api/models"
	acl2 "github.com/freshteapot/learnalist-api/server/pkg/acl"
)

// m exposing the data abstraction layer
type Manager struct {
	Datastore    models.Datastore
	Acl          acl.Acl
	Acl2         acl2.Acl
	DatabaseName string
	HugoHelper   hugo.HugoSiteBuilder
}

type HttpResponseMessage struct {
	Message string `json:"message"`
}

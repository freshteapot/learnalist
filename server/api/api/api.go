package api

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
)

// m exposing the data abstraction layer
type Manager struct {
	Datastore     models.Datastore
	Acl           acl.Acl
	DatabaseName  string
	HugoHelper    hugo.HugoSiteBuilder
	OauthHandlers oauth.Handlers
}

type HttpResponseMessage struct {
	Message string `json:"message"`
}

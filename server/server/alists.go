package server

import (
	authenticateAlists "github.com/freshteapot/learnalist-api/server/alists/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	alists "github.com/freshteapot/learnalist-api/server/alists/server"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"

	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
)

func InitAlists(acl acl.Acl, dal models.Datastore, hugoHelper *hugo.HugoHelper) {
	m := alists.Manager{
		Acl:             acl,
		Datastore:       dal,
		SiteCacheFolder: config.SiteCacheFolder,
		HugoHelper:      *hugoHelper,
	}

	authConfig := authenticate.Config{
		LookupBasic:  m.Datastore.UserWithUsernameAndPassword().Lookup,
		LookupBearer: m.Datastore.UserSession().GetUserUUIDByToken,
		Skip:         authenticateAlists.SkipAuth,
	}

	alists := server.Group("/alists")
	alists.Use(authenticate.Auth(authConfig))

	alists.GET("/*", m.GetAlist)
	server.Static("/", config.SiteCacheFolder)
}

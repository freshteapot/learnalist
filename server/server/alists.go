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

	server.GET("/logout.html", m.Logout)
	server.GET("/lists-by-me.html", m.GetMyLists, authenticate.Auth(authConfig))

	alists := server.Group("/alist")
	alists.Use(authenticate.Auth(authConfig))

	alists.GET("/*", m.GetAlist)

	// TODO http://localhost:1234/lists-by-me.html
	// TODO block access to the user files (alistsbyuser)

	server.Static("/", config.SiteCacheFolder)
}

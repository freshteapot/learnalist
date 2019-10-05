package server

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	alists "github.com/freshteapot/learnalist-api/server/alists/server"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/labstack/echo/v4/middleware"
)

func InitAlists(acl acl.Acl, dal models.Datastore, hugoHelper *hugo.HugoHelper) {
	m := alists.Manager{
		Acl:             acl,
		Acl2:            acl,
		Datastore:       dal,
		SiteCacheFolder: config.SiteCacheFolder,
		HugoHelper:      *hugoHelper,
	}

	authenticate.LookUp = m.Datastore.GetUserByCredentials

	alists := server.Group("/alists")
	alists.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper:   authenticate.SkipBasicAuth,
		Validator: authenticate.ValidateBasicAuth,
	}))

	alists.GET("/*", m.GetAlist)
	server.Static("/", config.SiteCacheFolder)
}

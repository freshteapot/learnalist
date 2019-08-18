package server

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	alists "github.com/freshteapot/learnalist-api/server/alists/server"
	"github.com/freshteapot/learnalist-api/server/api/acl"
)

func InitAlists(acl *acl.Acl, hugoHelper *hugo.HugoHelper) {
	m := alists.Manager{
		Acl:             *acl,
		SiteCacheFolder: config.SiteCacheFolder,
		HugoHelper:      *hugoHelper,
	}

	alists := server.Group("/alists")

	alists.GET("/*", m.GetAlist)
}

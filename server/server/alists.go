package server

import (
	alists "github.com/freshteapot/learnalist-api/server/alists/server"
	"github.com/freshteapot/learnalist-api/server/api/acl"
)

func InitAlists(acl *acl.Acl) {
	m := alists.Manager{
		Acl:              *acl,
		StaticSiteFolder: config.StaticSiteFolder,
	}
	alists.InitListenForFiles()
	alists := server.Group("/alists")

	alists.GET("/*", m.GetAlist)
}

package server

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	alists "github.com/freshteapot/learnalist-api/server/alists/server"
	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/robfig/cron"
)

func InitAlists(cron *cron.Cron, acl *acl.Acl, hugoHelper *hugo.HugoHelper) {
	cron.AddJob("@every 1s", hugoHelper.RegisterCron())

	m := alists.Manager{
		Acl:             *acl,
		SiteCacheFolder: config.SiteCacheFolder,
		HugoHelper:      *hugoHelper,
	}

	alists := server.Group("/alists")

	alists.GET("/*", m.GetAlist)
}

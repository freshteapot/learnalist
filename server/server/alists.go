package server

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	alists "github.com/freshteapot/learnalist-api/server/alists/server"
	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/robfig/cron"
)

func InitAlists(cron *cron.Cron, acl *acl.Acl) {
	cron.AddJob("@every 1s", hugo.RegisterContentCreationJob(hugo.JobConfig{
		StaticSiteFolder: config.StaticSiteFolder,
	}))

	m := alists.Manager{
		Acl:             *acl,
		SiteCacheFolder: config.SiteCacheFolder,
	}

	alists := server.Group("/alists")

	alists.GET("/*", m.GetAlist)
}

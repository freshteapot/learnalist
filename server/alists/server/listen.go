package server

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/robfig/cron"
)

func InitListenForFiles() {
	c := cron.New()
	c.Start()
	c.AddJob("@every 1s", hugo.RegisterContentCreationJob())
}

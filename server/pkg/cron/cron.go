package cron

import (
	"github.com/robfig/cron/v3"
)

func NewCron() *cron.Cron {
	masterCron := cron.New()
	masterCron.Start()
	return masterCron
}

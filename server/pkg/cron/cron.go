package cron

import (
	"github.com/robfig/cron"
)

func NewCron() *cron.Cron {
	masterCron := cron.New()
	masterCron.Start()
	return masterCron
}

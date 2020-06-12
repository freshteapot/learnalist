package hugo

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func (d Job) Run() {
	d.Helper.ProcessContent()
}

func (h HugoHelper) RegisterCronJob() {
	if h.externalHugo {
		fmt.Println("Will not process request as external hugo enabled")
		return
	}

	// Have a way to skip if hugoRunning as its own process / service
	if *h.cronEntryID != 0 {
		return
	}

	entryID, _ := h.cron.AddJob("@every 1s", Job{
		Helper: &h,
	})

	*h.cronEntryID = entryID
}

func (h HugoHelper) StopCronJob(logContext *logrus.Entry) {
	if *h.cronEntryID != 0 {
		logContext.WithFields(logrus.Fields{
			"event": "cron-stop",
		}).Info("done")
		h.cron.Remove(*h.cronEntryID)
		*h.cronEntryID = 0
	}
}

func (h HugoHelper) ProcessContent() {
	h.inprogress.Lock()
	logContext := h.logger.WithFields(logrus.Fields{
		"context": "hugo-build",
		"event":   "process-content",
	})
	logContext.Info("started")
	h.Build(logContext)
	h.inprogress.Unlock()
	logContext.Info("finished")
}

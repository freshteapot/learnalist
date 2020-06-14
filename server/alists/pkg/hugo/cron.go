package hugo

import (
	"github.com/sirupsen/logrus"
)

func (d Job) Run() {
	d.Helper.ProcessContent()
}

func (h HugoHelper) RegisterCronJob() {
	logContext := h.logger.WithFields(logrus.Fields{
		"context": "hugo-register",
	})

	if h.externalHugo {
		logContext.Info("skipping, hugo.external enabled")
		return
	}

	// Have a way to skip if hugoRunning as its own process / service
	if *h.cronEntryID != 0 {
		logContext.Info("skipping, cron already scheduled")
		return
	}

	*h.cronEntryID, _ = h.cron.AddJob("@every 1s", Job{
		Helper: h,
	})

	logContext.Info("scheduled")
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

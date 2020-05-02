package hugo

import (
	"fmt"
	"time"
)

type Job struct {
	Helper *HugoHelper
}

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

func (h HugoHelper) StopCronJob() {
	if *h.cronEntryID != 0 {
		fmt.Println("stopping cronjob")
		h.cron.Remove(*h.cronEntryID)
		*h.cronEntryID = 0
	}
}

func (h HugoHelper) ProcessContent() {
	h.inprogress.Lock()
	now := time.Now()
	fmt.Printf("Processing content within %s @ %s\n", h.Cwd, now)

	h.Build()
	h.inprogress.Unlock()
}

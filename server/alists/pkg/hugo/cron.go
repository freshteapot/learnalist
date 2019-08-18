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

func (h HugoHelper) RegisterCron() Job {
	return Job{
		Helper: &h,
	}
}

func (h HugoHelper) ProcessContent() {
	now := time.Now()
	fmt.Printf("Processing content within %s @ %s\n", h.Cwd, now)
	h.MakeContent()
	h.Build()
}

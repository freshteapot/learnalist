package hugo

import (
	"fmt"
	"time"
)

type Job struct{}

func (d Job) Run() {
	ProcessContent()
}

func RegisterContentCreationJob() Job {
	return Job{}
}

func ProcessContent() {
	now := time.Now()
	fmt.Printf("Processing content @ %s\n", now)
}

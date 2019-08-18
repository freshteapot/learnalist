package hugo

import (
	"fmt"
	"time"
)

type JobConfig struct {
	StaticSiteFolder string
}

type Job struct {
	Config JobConfig
}

func (d Job) Run() {
	ProcessContent(d.Config)
}

func RegisterContentCreationJob(config JobConfig) Job {
	return Job{
		Config: config,
	}
}

func ProcessContent(config JobConfig) {
	now := time.Now()
	fmt.Printf("Processing content within %s @ %s\n", config.StaticSiteFolder, now)
}

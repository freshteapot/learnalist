package hugo

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/robfig/cron/v3"
)

type HugoSiteBuilder interface {
	MakeContent()
	Build()
	Write(aList *alist.Alist)
}

type HugoHelper struct {
	Cwd              string
	DataDirectory    string
	ContentDirectory string
	cronEntryID      cron.EntryID
	cron             *cron.Cron
}

func NewHugoHelper(cwd string, _cron *cron.Cron) *HugoHelper {
	// TODO make sure the dataDir exists
	dataDirectory := fmt.Sprintf("%s/data/lists", cwd)
	contentDirectory := fmt.Sprintf("%s/content/alists", cwd)

	return &HugoHelper{
		Cwd:              cwd,
		DataDirectory:    dataDirectory,
		ContentDirectory: contentDirectory,
		cronEntryID:      0,
		cron:             _cron,
	}
}

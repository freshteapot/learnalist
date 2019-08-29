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
	// TODO should we remove via here?
}

type HugoHelper struct {
	Cwd              string
	DataDirectory    string
	ContentDirectory string
	cronEntryID      *cron.EntryID
	cron             *cron.Cron
}

func NewHugoHelper(cwd string, _cron *cron.Cron) *HugoHelper {
	// TODO make sure the dataDir exists
	dataDirectory := fmt.Sprintf("%s/data/lists", cwd)
	contentDirectory := fmt.Sprintf("%s/content/alists", cwd)
	// This is required to keep track of the memory, I think.
	var empty cron.EntryID
	empty = 0

	return &HugoHelper{
		Cwd:              cwd,
		DataDirectory:    dataDirectory,
		ContentDirectory: contentDirectory,
		cronEntryID:      &empty,
		cron:             _cron,
	}
}

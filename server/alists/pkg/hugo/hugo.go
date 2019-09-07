package hugo

import (
	"fmt"
	"log"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
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
	SiteCacheFolder  string
	cronEntryID      *cron.EntryID
	cron             *cron.Cron
}

func NewHugoHelper(cwd string, _cron *cron.Cron, siteCacheFolder string) *HugoHelper {
	// TODO maybe make a test run
	dataDirectory := fmt.Sprintf("%s/data/lists", cwd)
	if !utils.IsDir(dataDirectory) {
		log.Fatal(fmt.Sprintf("%s is not a directory", dataDirectory))
	}
	contentDirectory := fmt.Sprintf("%s/content/alists", cwd)
	if !utils.IsDir(contentDirectory) {
		log.Fatal(fmt.Sprintf("%s is not a directory", contentDirectory))
	}

	if !utils.IsDir(siteCacheFolder) {
		log.Fatal(fmt.Sprintf("%s is not a directory", siteCacheFolder))
	}
	// This is required to keep track of the memory, I think.
	var empty cron.EntryID
	empty = 0

	return &HugoHelper{
		Cwd:              cwd,
		DataDirectory:    dataDirectory,
		ContentDirectory: contentDirectory,
		SiteCacheFolder:  siteCacheFolder,
		cronEntryID:      &empty,
		cron:             _cron,
	}
}

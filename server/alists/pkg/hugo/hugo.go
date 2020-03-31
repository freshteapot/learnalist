package hugo

import (
	"fmt"
	"log"
	"sync"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/robfig/cron/v3"
)

type HugoSiteBuilder interface {
	ProcessContent()
	MakeContent()
	Build()
	WriteList(aList *alist.Alist)
	WriteListsByUser(userUUID string, lists []string)
	// Remove list via uuid
	Remove(uuid string)
}

type HugoHelper struct {
	Cwd              string
	DataDirectory    string
	ContentDirectory string
	SiteCacheFolder  string
	PublishDirectory string
	cronEntryID      *cron.EntryID
	cron             *cron.Cron
	inprogress       *sync.Mutex
}

func NewHugoHelper(cwd string, _cron *cron.Cron, siteCacheFolder string) *HugoHelper {
	// TODO maybe make a test run
	dataDirectory := fmt.Sprintf("%s/data/alist", cwd)
	if !utils.IsDir(dataDirectory) {
		log.Fatal(fmt.Sprintf("%s is not a directory", dataDirectory))
	}
	contentDirectory := fmt.Sprintf("%s/content/alist", cwd)
	if !utils.IsDir(contentDirectory) {
		log.Fatal(fmt.Sprintf("%s is not a directory", contentDirectory))
	}

	publishDirectory := fmt.Sprintf("%s/public", cwd)
	if !utils.IsDir(publishDirectory) {
		log.Fatal(fmt.Sprintf("%s is not a directory", publishDirectory))
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
		PublishDirectory: publishDirectory,
		cronEntryID:      &empty,
		cron:             _cron,
		inprogress:       &sync.Mutex{},
	}
}

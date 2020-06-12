package hugo

import (
	"sync"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type HugoSiteBuilder interface {
	ProcessContent()
	Build(logContext *logrus.Entry)
	WriteList(aList alist.Alist)
	WriteListsByUser(userUUID string, lists []alist.ShortInfo)
	WritePublicLists(lists []alist.ShortInfo)
	// Remove list via uuid
	DeleteList(uuid string) error
	DeleteUser(uuid string) error
}

type HugoHelper struct {
	cwd                string
	environment        string
	externalHugo       bool
	cronEntryID        *cron.EntryID
	cron               *cron.Cron
	inprogress         *sync.Mutex
	AlistWriter        HugoAListWriter
	AlistsByUserWriter HugoAListUserWriter
	PublicListsWriter  HugoPublicListsWriter
	logger             *logrus.Logger
}

type Job struct {
	Helper *HugoHelper
}

type HugoAListUserWriter struct {
	dataDirectory    string
	contentDirectory string
	publishDirectory string
}

type HugoAListWriter struct {
	dataDirectory    string
	contentDirectory string
	publishDirectory string
}

type HugoPublicListsWriter struct {
	dataDirectory string
}

const (
	RealtivePathData                      = "%s/data"
	RealtivePathContentAlist              = "%s/content/alist"
	RealtivePathDataAlist                 = "%s/data/alist"
	RealtivePathContentAlistsByUser       = "%s/content/alistsbyuser"
	RealtivePathDataAlistsByUser          = "%s/data/alistsbyuser"
	RealtivePathPublic                    = "%s/public"
	RealtivePathPublicContentAlist        = "%s/public/alist"
	RealtivePathPublicContentAlistsByUser = "%s/public/alistsbyuser"
)

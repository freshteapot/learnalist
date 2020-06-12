package hugo

import (
	"fmt"
	"log"
	"sync"

	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

var entryID cron.EntryID

func NewHugoHelper(cwd string, environment string, isExternal bool, _cron *cron.Cron, logger *logrus.Logger) HugoHelper {
	check := []string{
		RealtivePathContentAlist,
		RealtivePathDataAlist,
		RealtivePathContentAlistsByUser,
		RealtivePathDataAlistsByUser,
		RealtivePathPublic,
		RealtivePathPublicContentAlist,
		RealtivePathPublicContentAlistsByUser,
	}

	for _, template := range check {
		directory := fmt.Sprintf(template, cwd)
		if !utils.IsDir(directory) {
			log.Fatal(fmt.Sprintf("%s is not a directory", directory))
		}
	}

	// This is required to keep track of the memory, I think.
	var empty cron.EntryID
	empty = 0
	entryID = 0
	publishDirectory := fmt.Sprintf(RealtivePathPublic, cwd)

	return HugoHelper{
		logger:       logger,
		cwd:          cwd,
		environment:  environment,
		externalHugo: isExternal,
		cronEntryID:  &empty,
		cron:         _cron,
		inprogress:   &sync.Mutex{},
		AlistWriter: NewHugoAListWriter(
			fmt.Sprintf(RealtivePathContentAlist, cwd),
			fmt.Sprintf(RealtivePathDataAlist, cwd),
			publishDirectory),
		AlistsByUserWriter: NewHugoAListByUserWriter(
			fmt.Sprintf(RealtivePathContentAlistsByUser, cwd),
			fmt.Sprintf(RealtivePathDataAlistsByUser, cwd),
			publishDirectory),
		PublicListsWriter: NewHugoPublicListsWriter(fmt.Sprintf(RealtivePathData, cwd)),
	}
}

func (h HugoHelper) GetPubicDirectory() string {
	return fmt.Sprintf(RealtivePathPublic, h.cwd)
}

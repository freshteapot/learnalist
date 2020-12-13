package hugo

import (
	"fmt"
	"log"
	"sync"

	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

func NewHugoHelper(cwd string, environment string, isExternal bool, _cron *cron.Cron, logger *logrus.Logger) HugoHelper {
	check := []string{
		RealtivePathContentAlist,
		RealtivePathDataAlist,
		RealtivePathContentAlistsByUser,
		RealtivePathDataAlistsByUser,
		RealtivePathPublic,
		RealtivePathPublicContentAlist,
		RealtivePathPublicContentAlistsByUser,
		RealtivePathChallengeContent,
		RealtivePathChallengeData,
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
	writer := NewHugoFileWriter(logger.WithField("context", "hugo-writer"))

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
			publishDirectory,
			writer),
		AlistsByUserWriter: NewHugoAListByUserWriter(
			fmt.Sprintf(RealtivePathContentAlistsByUser, cwd),
			fmt.Sprintf(RealtivePathDataAlistsByUser, cwd),
			publishDirectory,
			writer),
		PublicListsWriter: NewHugoPublicListsWriter(
			fmt.Sprintf(RealtivePathData, cwd),
			publishDirectory,
			writer),
		challengeWriter: NewChallengeWriter(
			fmt.Sprintf(RealtivePathChallengeContent, cwd),
			fmt.Sprintf(RealtivePathChallengeData, cwd),
			publishDirectory,
			writer,
		),
	}
}

func (h HugoHelper) GetPubicDirectory() string {
	return fmt.Sprintf(RealtivePathPublic, h.cwd)
}

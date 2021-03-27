package hugo

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

func NewHugoHelper(cwd string, environment string, logger logrus.FieldLogger) *HugoHelper {
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

	writer := NewHugoFileWriter(logger.WithField("context", "hugo-writer"))

	publishDirectory := fmt.Sprintf(RealtivePathPublic, cwd)
	return &HugoHelper{
		contentWillBuildTimer: nil,
		logContext:            logger,
		cwd:                   cwd,
		environment:           environment,
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

func (h *HugoHelper) GetPubicDirectory() string {
	return fmt.Sprintf(RealtivePathPublic, h.cwd)
}

func (h *HugoHelper) Subscribe(topic string, sc stan.Conn) error {
	var err error
	handle := func(msg *stan.Msg) {
		var moment event.Eventlog
		json.Unmarshal(msg.Data, &moment)
		h.OnEvent(moment)
	}

	durableName := "static-site-hugo"
	h.subscription, err = sc.Subscribe(
		topic,
		handle,
		stan.DurableName(durableName),
		stan.DeliverAllAvailable(),
		stan.MaxInflight(1),
	)
	if err == nil {
		h.logContext.Info("Running")
	}
	return err
}

func (h *HugoHelper) Close() {
	err := h.subscription.Close()
	if err != nil {
		h.logContext.WithField("error", err).Error("closing subscription")
	}
}

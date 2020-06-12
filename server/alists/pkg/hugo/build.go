package hugo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

func (h HugoHelper) Build(logContext *logrus.Entry) {
	a := h.AlistWriter.GetFilesToPublish()
	b := h.AlistsByUserWriter.GetFilesToPublish()

	logContext.WithFields(logrus.Fields{
		"event":      "build-stats",
		"lists":      len(a),
		"user_lists": len(b),
	}).Info("stats")

	toPublish := append(a, b...)

	if len(toPublish) == 0 {
		logContext.WithFields(logrus.Fields{
			"event": "no-content",
		}).Info("Nothing to publish")
		h.StopCronJob(logContext)
		return
	}

	h.buildSite(logContext)
	h.StopCronJob(logContext)
}

func (h HugoHelper) buildSite(logContext *logrus.Entry) {
	logContext = logContext.WithFields(logrus.Fields{
		"event": "build-site",
	})

	staticSiteFolder := h.cwd
	parts := []string{
		"-verbose",
		fmt.Sprintf(`--environment=%s`, h.environment),
	}

	cmd := exec.Command("hugo", parts...)
	cmd.Dir = staticSiteFolder
	out, err := cmd.Output()
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"out":   string(out),
		}).Fatal("failed")
	}

	logContext.WithFields(logrus.Fields{
		"out": string(out),
	}).Info("done")
}

func (h HugoHelper) deleteFiles(files []string) {
	log := h.logger
	logContext := log.WithFields(logrus.Fields{
		"event": "delete-file",
	})

	// Command to remove lists in hugo that are no longer in the DB
	// https://github.com/freshteapot/learnalist-api/issues/98
	for _, path := range files {
		err := os.Remove(path)
		if err != nil {
			if !strings.HasSuffix(err.Error(), "no such file or directory") {
				logContext.WithFields(logrus.Fields{
					"path":  path,
					"error": err,
				}).Error("file removed")
				continue
			}
		}

		logContext.WithFields(logrus.Fields{
			"path": path,
		}).Info("file removed")
	}
}

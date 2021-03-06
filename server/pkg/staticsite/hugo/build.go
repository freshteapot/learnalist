package hugo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func (h *HugoHelper) ProcessContent() {
	// Protecting against a herd of events
	// Controlling how often we rebuild
	if h.contentWillBuildTimer != nil {
		return
	}

	h.contentWillBuildTimer = time.AfterFunc(500*time.Millisecond, func() {
		logContext := h.logContext.WithFields(logrus.Fields{
			"context": "hugo-build",
			"event":   "process-content",
		})

		logContext.Info("started")
		h.Build(logContext)
		logContext.Info("finished")
		h.contentWillBuildTimer = nil
	})
}

func (h *HugoHelper) Build(logContext *logrus.Entry) {
	a := h.AlistWriter.GetFilesToPublish()
	b := h.AlistsByUserWriter.GetFilesToPublish()
	c := h.challengeWriter.GetFilesToPublish()

	toPublish := a
	toPublish = append(toPublish, b...)
	toPublish = append(toPublish, c...)
	// Trigger a rebuild regardless

	logContext.WithFields(logrus.Fields{
		"event": "build-stats",
		"stats": map[string]interface{}{
			"lists":      len(a),
			"user_lists": len(b),
			"challenges": len(c),
		},
	}).Info("stats")

	err := h.buildSite(logContext)
	if err != nil {
		err := h.buildSite(logContext)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"event": "repeat-attempt-failed",
				"error": err,
			}).Error("failed building hugo")
		}
	}

	h.deleteFiles(toPublish)
}

func (h *HugoHelper) buildSite(logContext *logrus.Entry) error {
	logContext = logContext.WithFields(logrus.Fields{
		"event": "build-site",
	})

	staticSiteFolder := h.cwd
	parts := []string{
		"-verbose",
		fmt.Sprintf(`--environment=%s`, h.environment),
		"--quiet",
		"--ignoreCache",
	}

	cmd := exec.Command("hugo", parts...)

	cmd.Dir = staticSiteFolder
	out, err := cmd.Output()
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"out":   string(out),
		}).Error("failed")
		return err
	}

	//logContext.WithFields(logrus.Fields{
	//	"out": string(out),
	//}).Info("done")
	return nil
}

func (h *HugoHelper) deleteFiles(files []string) {
	logContext := h.logContext.WithFields(logrus.Fields{
		"event": "delete-file",
	})

	// Not smart enough to remove data, come back to this
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

		//logContext.WithFields(logrus.Fields{
		//	"path": path,
		//}).Info("file removed")
	}
}

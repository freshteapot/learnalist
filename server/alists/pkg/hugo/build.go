package hugo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
)

func (h HugoHelper) Build() {
	a := h.AlistWriter.GetFilesToPublish()
	b := h.AlistsByUserWriter.GetFilesToPublish()

	fmt.Printf("Build static site for %d lists\n", len(a))
	fmt.Printf("Build static site for %d my lists\n", len(b))

	toPublish := append(a, b...)

	if len(toPublish) == 0 {
		fmt.Println("Nothing to publish")
		h.StopCronJob()
		return
	}

	h.buildSite()
	h.StopCronJob()
}

func (h HugoHelper) buildSite() {
	log := h.logger
	logContext := log.WithFields(logrus.Fields{
		"event": "build-site",
	})

	staticSiteFolder := h.Cwd
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
	// TODO Create an issue about a command line option to purge lists that are not in the database
	// Assume one day, this will get out of sync.
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

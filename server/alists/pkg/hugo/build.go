package hugo

import (
	"fmt"
	"log"
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
	staticSiteFolder := h.Cwd
	// TODO change this to be dynamic via config
	parts := []string{
		"-verbose",
		fmt.Sprintf(`--environment=%s`, h.Environment),
	}

	cmd := exec.Command("hugo", parts...)
	cmd.Dir = staticSiteFolder
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(string(out))
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

func (h HugoHelper) deleteFiles(files []string) {
	log := h.logger
	// TODO Create an issue about a command line option to purge lists that are not in the database
	// Assume one day, this will get out of sync.
	for _, path := range files {
		err := os.Remove(path)
		if err != nil {
			if !strings.HasSuffix(err.Error(), "no such file or directory") {
				log.WithFields(logrus.Fields{
					"event": "delete-file",
					"path":  path,
					"err":   err,
				}).Error("file removed")
				continue
			}
		}

		log.WithFields(logrus.Fields{
			"event": "delete-file",
			"path":  path,
		}).Info("file removed")
	}
}

package hugo

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
)

func (h HugoHelper) Build() {
	toPublish := h.getFilesToPublish()
	if len(toPublish) == 0 {
		fmt.Println("Nothing to publish")
		h.StopCronJob()
		return
	}

	fmt.Printf("Build static site for %d lists\n", len(toPublish))
	h.buildSite()
	uuids := h.getPublishedFiles()

	h.copyToSiteCache()

	// Only remove what we processed, that way any that get added will not be lost (hopefully)
	for _, uuid := range uuids {
		h.deleteBuildFiles(uuid)
	}
}

func (h HugoHelper) buildSite() {
	staticSiteFolder := h.Cwd
	parts := strings.Split("-e alist --config=config/alist/config.toml", " ")
	cmd := exec.Command("hugo", parts...)
	cmd.Dir = staticSiteFolder
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

// Copy each file over, including non alist files and directories
func (h HugoHelper) copyToSiteCache() {
	siteCacheDir := h.SiteCacheFolder
	destinationDir := h.PublishDirectory

	err := copy.Copy(destinationDir, siteCacheDir)
	fmt.Println(err)
}

func (h HugoHelper) getFilesToPublish() []string {
	dataDir := h.ContentDirectory

	var files []string
	var uuids []string
	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		filename := strings.TrimPrefix(file, dataDir+"/")

		if strings.HasSuffix(filename, ".md") {
			uuid := strings.TrimSuffix(filename, ".md")
			uuids = append(uuids, uuid)
		}
	}
	return uuids
}

func (h HugoHelper) getPublishedFiles() []string {
	dataDir := h.PublishDirectory
	var files []string
	var uuids []string
	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, file := range files {
		filename := strings.TrimPrefix(file, dataDir+"/")

		if strings.HasPrefix(filename, "alists/") && strings.HasSuffix(filename, ".html") {
			uuid := strings.TrimPrefix(filename, "alists/")
			uuid = strings.TrimSuffix(uuid, ".html")
			uuids = append(uuids, uuid)
		}
	}
	return uuids
}

// deleteFiles
// 	- Remove from content
// 	- Remove from data directory
//	- Remove from public-alist directory

func (h HugoHelper) getBuildFiles(uuid string) []string {
	files := []string{
		fmt.Sprintf("%s/%s.md", h.ContentDirectory, uuid),
		fmt.Sprintf("%s/%s.json", h.DataDirectory, uuid),
		fmt.Sprintf("%s/alists/%s.json", h.PublishDirectory, uuid),
		fmt.Sprintf("%s/alists/%s.html", h.PublishDirectory, uuid),
	}
	return files
}

func (h HugoHelper) deleteBuildFiles(uuid string) {
	files := h.getBuildFiles(uuid)
	h.deleteFiles(files)
}

func (h HugoHelper) deleteFiles(files []string) {
	for _, path := range files {
		fmt.Printf("Removing %s\n", path)
		err := os.Remove(path)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to remove %s", err))
		}
	}
}

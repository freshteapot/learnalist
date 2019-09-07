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
		h.deleteFiles(uuid)
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
	staticSiteFolder := h.Cwd
	siteCacheDir := h.SiteCacheFolder
	destinationDir := staticSiteFolder + "/public-alist/"

	err := copy.Copy(destinationDir, siteCacheDir)
	fmt.Println(err)
}

func (h HugoHelper) getFilesToPublish() []string {
	staticSiteFolder := h.Cwd
	dataDir := staticSiteFolder + "/content/alists"

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
	staticSiteFolder := h.Cwd
	dataDir := staticSiteFolder + "/public-alist"
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
func (h HugoHelper) deleteFiles(uuid string) {
	staticSiteFolder := h.Cwd
	files := []string{
		fmt.Sprintf("%s/content/alists/%s.md", staticSiteFolder, uuid),
		fmt.Sprintf("%s/data/lists/%s.json", staticSiteFolder, uuid),
		fmt.Sprintf("%s/public-alist/alists/%s.json", staticSiteFolder, uuid),
		fmt.Sprintf("%s/public-alist/alists/%s.html", staticSiteFolder, uuid),
	}

	for _, path := range files {
		fmt.Printf("Removing %s\n", path)
		err := os.Remove(path)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to remove %s", err))
		}
	}
}

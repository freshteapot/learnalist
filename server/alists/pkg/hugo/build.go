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
	fmt.Println(toPublish)
	h.buildSite()
	//uuids := h.getPublishedFiles()
	uuids := toPublish

	h.copyToSiteCache()

	// Only remove what we processed, that way any that get added will not be lost (hopefully)
	for _, uuid := range uuids {
		h.deleteBuildFiles(uuid)
	}
}

func (h HugoHelper) buildSite() {
	staticSiteFolder := h.Cwd
	// TODO change this to be dynamic via config
	parts := []string{
		"-verbose",
		`--environment=production`,
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
		fmt.Print("Something has gone wrong in getFilesToPublish, when looking for files to process")
		fmt.Println(len(files))
		fmt.Println(err)
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

// TODO change this to lists
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
		fmt.Print("Something has gone wrong in getPublishedFiles, when looking for files to process")
		fmt.Println(len(files))
		fmt.Println(err)
	}

	for _, file := range files {
		filename := strings.TrimPrefix(file, dataDir+"/")
		fmt.Println(file)
		if strings.HasPrefix(filename, "alist/") && strings.HasSuffix(filename, ".html") {
			uuid := strings.TrimPrefix(filename, "alist/")
			uuid = strings.TrimSuffix(uuid, ".html")
			uuids = append(uuids, uuid)
		}
	}
	return uuids
}

func (h HugoHelper) getBuildFiles(uuid string) []string {
	files := []string{
		fmt.Sprintf("%s/%s.md", h.ContentDirectory, uuid),
		fmt.Sprintf("%s/%s.json", h.DataDirectory, uuid),
		fmt.Sprintf("%s/alist/%s.json", h.PublishDirectory, uuid),
		fmt.Sprintf("%s/alist/%s.html", h.PublishDirectory, uuid),
	}
	return files
}

// deleteBuildFiles
// 	- Remove from content
// 	- Remove from data directory
//	- Remove from hugo publishe directory
func (h HugoHelper) deleteBuildFiles(uuid string) {
	files := h.getBuildFiles(uuid)
	h.deleteFiles(files)
}

func (h HugoHelper) deleteFiles(files []string) {
	// TODO Create an issue about a command line option to purge lists that are not in the database
	// Assume one day, this will get out of sync.
	for _, path := range files {
		fmt.Printf("Removing %s\n", path)
		err := os.Remove(path)
		if err != nil {
			if !strings.HasSuffix(err.Error(), "no such file or directory") {
				fmt.Println(fmt.Sprintf("Failed to remove %s", err))
			}
		}
	}
}

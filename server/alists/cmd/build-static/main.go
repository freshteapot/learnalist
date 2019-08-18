package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	staticSiteFolder := flag.String("static", "", "path to static site builder")

	flag.Parse()
	*staticSiteFolder = strings.TrimRight(*staticSiteFolder, "/")

	toPublish := getFilesToPublish(*staticSiteFolder)
	if len(toPublish) == 0 {
		fmt.Println("Nothing to publish")
		return
	}

	fmt.Printf("Build static site for %d lists\n", len(toPublish))
	buildSite(*staticSiteFolder)
	uuids := getPublishedFiles(*staticSiteFolder)

	// Copy each file over, including non alist files
	copyToSiteCache(*staticSiteFolder)

	// Empty hugo destination dir
	emptyDestinationDir(*staticSiteFolder)

	// Only remove what we processed, that way any that get added will not be lost (hopefully)
	for _, uuid := range uuids {
		// Remove from content
		// Remove from data directory
		deleteFiles(*staticSiteFolder, uuid)
	}
}

func buildSite(staticSiteFolder string) {
	parts := strings.Split("--cleanDestinationDir -e alist --config=config/alist/config.toml", " ")
	cmd := exec.Command("hugo", parts...)
	cmd.Dir = staticSiteFolder
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

func copyToSiteCache(staticSiteFolder string) {
	destinationDir := "./public-alist/"
	siteCacheDir := "../site-cache/"

	parts := []string{"-r", destinationDir, siteCacheDir}
	cmd := exec.Command("cp", parts...)
	cmd.Dir = staticSiteFolder
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

func emptyDestinationDir(staticSiteFolder string) {
	destinationDir := "./public-alist/"
	parts := []string{"-r", destinationDir}
	cmd := exec.Command("rm", parts...)
	cmd.Dir = staticSiteFolder
	out, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

func getFilesToPublish(staticSiteFolder string) []string {
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

func getPublishedFiles(staticSiteFolder string) []string {
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

func deleteFiles(staticSiteFolder string, uuid string) {
	files := []string{
		fmt.Sprintf("%s/content/alists/%s.md", staticSiteFolder, uuid),
		fmt.Sprintf("%s/data/lists/%s.json", staticSiteFolder, uuid),
	}

	for _, path := range files {
		fmt.Printf("Removing %s\n", path)
		err := os.Remove(path)
		if err != nil {
			fmt.Println(err)
		}
	}
}

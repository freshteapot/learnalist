package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func doSingle(uuid string, dir string) {
	template := `
+++
Uuid = "%s"
+++
`
	content := strings.TrimSpace(fmt.Sprintf(template, uuid))

	path := fmt.Sprintf("%s/%s.md", dir, uuid)
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func doAll(dataDir string, contentDir string) {
	var files []string
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
		uuid := strings.TrimSuffix(filename, ".json")
		doSingle(uuid, contentDir)
	}
}

func main() {
	// Not overly safe
	dataDir := flag.String("data-dir", "./data/lists", "data directory for the lists")
	contentDir := flag.String("content-dir", "./content/alists", "content directory for the lists")
	uuid := flag.String("uuid", "", "alist uuid")
	flag.Parse()

	// # TODO check that dataDir exists.
	// # TODO check that contentDir exists.
	*dataDir = strings.TrimRight(*dataDir, "/")
	*contentDir = strings.TrimRight(*contentDir, "/")

	if *uuid != "" {
		doSingle(*uuid, *contentDir)
	} else {
		doAll(*dataDir, *contentDir)
	}
}

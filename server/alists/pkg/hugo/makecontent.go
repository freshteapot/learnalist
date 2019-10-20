package hugo

import (
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

func (h HugoHelper) MakeContent() {
	dataDir := h.DataDirectory
	contentDir := h.ContentDirectory

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

	// TODO I suspect, I can ignore this
	if err != nil {
		fmt.Println(len(files))
		fmt.Println(err)
		panic(err)
	}

	for _, file := range files {
		filename := strings.TrimPrefix(file, dataDir+"/")
		uuid := strings.TrimSuffix(filename, ".json")
		doSingle(uuid, contentDir)
	}
}

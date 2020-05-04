package hugo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

type HugoAListWriter struct {
	dataDirectory    string
	contentDirectory string
	publishDirectory string
}

func NewHugoAListWriter(contentDirectory string, dataDirectory string, publishDirectory string) HugoAListWriter {
	return HugoAListWriter{
		dataDirectory:    dataDirectory,
		contentDirectory: contentDirectory,
		publishDirectory: publishDirectory,
	}
}

func (w HugoAListWriter) Content(aList alist.Alist) {
	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"stringify_string_array": func(input []string) string {
			urlsJson, _ := json.Marshal(input)
			return string(urlsJson)
		},

		"can_interact": func(interact *alist.Interact) bool {
			if interact.Slideshow == "1" {
				return true
			}
			return false
		},

		"js_include": func(info alist.AlistInfo) string {
			jsInclude := make([]string, 0)
			jsInclude = append(jsInclude, info.ListType)
			b, _ := json.Marshal(jsInclude)
			return string(b)
		},
	}

	base := template.Must(template.New("").Funcs(funcMap).Parse(`
---
uuid: {{.Uuid}}
title: {{.Info.Title}}
labels: {{stringify_string_array .Info.Labels}}
interact: {{can_interact .Info.Interact}}
js_include: {{js_include .Info}}
---
`))
	var tpl bytes.Buffer
	base.Execute(&tpl, aList)
	content := strings.TrimSpace(tpl.String())

	path := fmt.Sprintf("%s/%s.md", w.contentDirectory, aList.Uuid)

	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func (w HugoAListWriter) Data(aList alist.Alist) {
	uuid := aList.Uuid
	content, _ := json.Marshal(aList)
	path := fmt.Sprintf("%s/%s.json", w.dataDirectory, uuid)
	err := ioutil.WriteFile(path, content, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func (w HugoAListWriter) GetFilesToPublish() []string {
	var (
		files     []string
		toPublish []string
	)

	err := filepath.Walk(w.contentDirectory, func(path string, info os.FileInfo, err error) error {
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

	for _, path := range files {
		if strings.HasSuffix(path, ".md") {
			toPublish = append(toPublish, path)
		}
	}
	return toPublish
}

func (w HugoAListWriter) GetFilesToClean() []string {
	toPublish := w.GetFilesToPublish()
	var toClean []string

	for _, path := range toPublish {
		filename := strings.TrimPrefix(path, w.contentDirectory+"/")

		if strings.HasSuffix(filename, ".md") {
			uuid := strings.TrimSuffix(filename, ".md")

			filesToClean := []string{
				fmt.Sprintf("%s/%s.md", w.contentDirectory, uuid),
				fmt.Sprintf("%s/%s.json", w.dataDirectory, uuid),
				// TODO this might not be needed
				fmt.Sprintf("%s/alist/%s.json", w.publishDirectory, uuid),
				fmt.Sprintf("%s/alist/%s.html", w.publishDirectory, uuid),
			}
			toClean = append(toClean, filesToClean...)
		}
	}

	return toClean
}

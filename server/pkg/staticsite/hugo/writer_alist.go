package hugo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

func NewHugoAListWriter(contentDirectory string, dataDirectory string, publishDirectory string, writer FileWriter) HugoAListWriter {
	return HugoAListWriter{
		dataDirectory:    dataDirectory,
		contentDirectory: contentDirectory,
		publishDirectory: publishDirectory,
		writer:           writer,
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
			// Handle when it has not been set
			// Not all lists support interact
			if interact == nil {
				return false
			}

			if interact.Slideshow == 1 {
				return true
			}
			if interact.TotalRecall == 1 {
				return true
			}

			return false
		},

		"js_include": func(info alist.AlistInfo) string {
			include := make([]string, 0)
			include = append(include, "main")
			include = append(include, info.ListType)
			b, _ := json.Marshal(include)
			return string(b)
		},

		"css_include": func(info alist.AlistInfo) string {
			include := make([]string, 0)
			include = append(include, "main")
			include = append(include, info.ListType)
			b, _ := json.Marshal(include)
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
css_include: {{css_include .Info}}
---
`))
	var tpl bytes.Buffer
	base.Execute(&tpl, aList)
	content := strings.TrimSpace(tpl.String())

	path := fmt.Sprintf("%s/%s.md", w.contentDirectory, aList.Uuid)
	fmt.Println(path)
	w.writer.Write(path, []byte(content))
}

func (w HugoAListWriter) Data(aList alist.Alist) {
	uuid := aList.Uuid
	content, _ := json.Marshal(aList)
	path := fmt.Sprintf("%s/%s.json", w.dataDirectory, uuid)
	w.writer.Write(path, content)
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
		fmt.Print("Something has gone wrong in HugoAListWriter.getFilesToPublish, when looking for files to process")
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

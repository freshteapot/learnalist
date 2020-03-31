package hugo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

func doSingle(aList alist.Alist, dir string) {
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

	path := fmt.Sprintf("%s/%s.md", dir, aList.Uuid)

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

	if err != nil {
		fmt.Print("Something has gone wrong in MakeContent, when looking for files to process")
		fmt.Println(len(files))
		fmt.Println(err)
	}

	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		var aList alist.Alist
		aList.UnmarshalJSON(data)

		doSingle(aList, contentDir)
	}
}

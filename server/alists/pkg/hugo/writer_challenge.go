package hugo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
)

func NewChallengeWriter(contentDirectory string, dataDirectory string, publishDirectory string, writer FileWriter) hugoChallengeWriter {
	return hugoChallengeWriter{
		dataDirectory:    dataDirectory,
		contentDirectory: contentDirectory,
		publishDirectory: publishDirectory,
		writer:           writer,
	}
}

func (w hugoChallengeWriter) Content(entry challenge.ChallengeInfo) {
	base := template.Must(template.New("").Parse(`
---
uuid: {{.UUID}}
title: {{.Description}}
type: "challenge"
layout: {{.Kind}}
js_include: ["main"]
css_include: ["main"]
---
`))
	var tpl bytes.Buffer
	base.Execute(&tpl, entry)
	content := strings.TrimSpace(tpl.String())
	path := fmt.Sprintf("%s/%s.md", w.contentDirectory, entry.UUID)

	w.writer.Write(path, []byte(content))
}

func (w hugoChallengeWriter) Data(entry challenge.ChallengeInfo) {
	uuid := entry.UUID
	content, _ := json.Marshal(entry)
	path := fmt.Sprintf("%s/%s.json", w.dataDirectory, uuid)
	w.writer.Write(path, content)
}

func (w hugoChallengeWriter) Remove(uuid string) {
	w.writer.Remove(fmt.Sprintf("%s/%s.md", w.contentDirectory, uuid))
	w.writer.Remove(fmt.Sprintf("%s/%s.json", w.dataDirectory, uuid))
	w.writer.Remove(fmt.Sprintf("%s/%s.html", w.publishDirectory, uuid))
}

func (w hugoChallengeWriter) GetFilesToPublish() []string {
	var (
		files []string
	)

	err := filepath.Walk(w.contentDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		fmt.Print("Something has gone wrong in getFilesToPublish, when looking for files to process")
		fmt.Println(len(files))
		fmt.Println(err)
		files = make([]string, 0)
	}

	return files
}

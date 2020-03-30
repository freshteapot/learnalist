package tools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/spf13/cobra"
)

var hugoImportListsCmd = &cobra.Command{
	Use:   "hugo-import-lists",
	Short: "Convert JSON to hugo",
	RunE: func(cmd *cobra.Command, args []string) error {
		scanner := bufio.NewScanner(os.Stdin)
		contentDirectory, err := getFullPath(cmd.Flags().GetString("content-dir"))
		if err != nil {
			return err
		}

		dataDirectory, err := getFullPath(cmd.Flags().GetString("data-dir"))
		if err != nil {
			return err
		}
		fmt.Println(contentDirectory)
		fmt.Println(dataDirectory)

		writer := HugoAlist{
			contentDirectory: contentDirectory,
			dataDirectory:    dataDirectory,
		}

		for scanner.Scan() {
			input := scanner.Text()
			fmt.Println(input)
			var aList alist.Alist
			aList.UnmarshalJSON([]byte(input))

			writer.Data(aList)
			writer.Content(aList)
		}
		return nil
	},
}

type HugoAlist struct {
	dataDirectory    string
	contentDirectory string
}

func (writer HugoAlist) Content(aList alist.Alist) error {
	path := fmt.Sprintf("%s/alist/%s.md", writer.contentDirectory, aList.Uuid)
	template := writer.toContent(aList)
	return writeToFile(path, []byte(template))
}

func (writer HugoAlist) Data(aList alist.Alist) error {
	path := fmt.Sprintf("%s/alist/%s.json", writer.dataDirectory, aList.Uuid)
	b, _ := json.Marshal(aList)
	return writeToFile(path, b)
}

func (writer HugoAlist) toContent(aList alist.Alist) string {
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
	return strings.TrimSpace(tpl.String())
}

func init() {
	hugoImportListsCmd.Flags().String("content-dir", "", "Path to content dir")
	hugoImportListsCmd.Flags().String("data-dir", "", "Path to data dir")
}

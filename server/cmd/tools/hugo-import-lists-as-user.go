package tools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/spf13/cobra"
)

var hugoImportListsByUserCmd = &cobra.Command{
	Use:   "hugo-import-lists-by-user",
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

		UUID, _ := cmd.Flags().GetString("user-uuid")
		if UUID == "" {
			return errors.New("You are missing user uuid")
		}

		writer := HugoByUser{
			contentDirectory: contentDirectory,
			dataDirectory:    dataDirectory,
		}

		lists := make([]string, 0)
		for scanner.Scan() {
			input := scanner.Text()
			var aList alist.Alist
			err = aList.UnmarshalJSON([]byte(input))
			if err != nil {
				continue
			}
			lists = append(lists, aList.Uuid)
		}
		writer.Data(UUID, lists)
		writer.Content(UUID)
		return nil
	},
}

type HugoByUser struct {
	dataDirectory    string
	contentDirectory string
}

func (writer HugoByUser) Content(UUID string) error {
	path := fmt.Sprintf("%s/alistsbyuser/%s.md", writer.contentDirectory, UUID)
	template := writer.toContent(UUID)
	return writeToFile(path, []byte(template))
}

func (writer HugoByUser) Data(UUID string, lists []string) error {
	path := fmt.Sprintf("%s/alistsbyuser/%s.json", writer.dataDirectory, UUID)
	b, _ := json.Marshal(lists)
	return writeToFile(path, b)
}

func (writer HugoByUser) toContent(UUID string) string {
	data := struct {
		UUID string
	}{
		UUID: UUID,
	}

	base := template.Must(template.New("").Parse(`
---
title: "My Lists"
type: "alist"
layout: "user"
Uuid: {{.UUID}}
js_include: ["main"]
---
`))
	var tpl bytes.Buffer
	base.Execute(&tpl, data)
	return strings.TrimSpace(tpl.String())
}

func init() {
	hugoImportListsByUserCmd.Flags().String("content-dir", "", "Path to content dir")
	hugoImportListsByUserCmd.Flags().String("data-dir", "", "Path to data dir")
	hugoImportListsByUserCmd.Flags().String("user-uuid", "", "user uuid to link too")
}

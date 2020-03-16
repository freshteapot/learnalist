package tools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/spf13/cobra"
)

var hugoInputCmd = &cobra.Command{
	Use:   "hugo-input",
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

		for scanner.Scan() {
			input := scanner.Text()
			var aList alist.Alist
			aList.UnmarshalJSON([]byte(input))

			template := toContent(aList)
			writeData(dataDirectory, aList)
			writeContent(contentDirectory, template, aList)
		}
		return nil
	},
}

func getFullPath(input string, err error) (string, error) {
	if err != nil {
		return "", err
	}

	path := input
	if !utils.IsDir(path) {
		return "", errors.New(fmt.Sprintf("%s is not a directory", path))
	}

	return strings.TrimRight(path, "/"), nil
}

func writeContent(directory string, data string, aList alist.Alist) error {
	path := fmt.Sprintf("%s/%s.md", directory, aList.Uuid)
	template := toContent(aList)
	err := ioutil.WriteFile(path, []byte(template), 0644)
	if err != nil {
		return err
	}
	return nil
}

func writeData(directory string, aList alist.Alist) error {
	path := fmt.Sprintf("%s/%s.json", directory, aList.Uuid)

	b, _ := json.Marshal(aList)
	err := ioutil.WriteFile(path, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func toContent(aList alist.Alist) string {
	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"stringify_string_array": func(input []string) string {
			urlsJson, _ := json.Marshal(input)
			return string(urlsJson)
		},
	}

	base := template.Must(template.New("").Funcs(funcMap).Parse(`
---
Uuid: {{.Uuid}}
Title: {{.Info.Title}}
Labels: {{stringify_string_array .Info.Labels}}
---
`))
	var tpl bytes.Buffer
	base.Execute(&tpl, aList)
	return strings.TrimSpace(tpl.String())
}

func init() {
	hugoInputCmd.Flags().String("content-dir", "", "Path to content dir")
	hugoInputCmd.Flags().String("data-dir", "", "Path to data dir")
}

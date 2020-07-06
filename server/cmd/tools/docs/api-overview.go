package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
)

type Endpoint struct {
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Description string   `json:"description"`
	StatusCodes []string `json:"status_codes"`
}

type DocsApiOverview struct{}

func (d DocsApiOverview) GetEndpoints(input []byte) []Endpoint {
	var data map[string]interface{}
	json.Unmarshal(input, &data)
	query, err := gojq.Parse(".paths | to_entries")
	if err != nil {
		log.Fatalln(err)
	}

	var endpoints []Endpoint
	iter := query.Run(data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}

		for _, e := range v.([]interface{}) {
			e := e.(map[string]interface{})

			endpoint := d.parseEndpoint(e["key"].(string), e["value"].(map[string]interface{}))
			endpoints = append(endpoints, endpoint...)
		}
	}
	return endpoints
}

// parseEndpoint extract out path details and find all possible statuscodes
func (d DocsApiOverview) parseEndpoint(key string, value map[string]interface{}) []Endpoint {
	var endpoints []Endpoint
	query, _ := gojq.Parse("to_entries")
	iter := query.Run(value) // or query.RunWithContext
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}

		for _, e := range v.([]interface{}) {
			e := e.(map[string]interface{})
			responses := e["value"].(map[string]interface{})["responses"].(map[string]interface{})
			statusCodes := make([]string, 0)
			for k := range responses {
				statusCodes = append(statusCodes, k)
			}

			endpoint := Endpoint{
				Path:        key,
				Method:      e["key"].(string),
				Description: e["value"].(map[string]interface{})["description"].(string),
				StatusCodes: statusCodes,
			}
			endpoints = append(endpoints, endpoint)
		}
	}
	return endpoints
}

func (d DocsApiOverview) Render(endpoints []Endpoint) string {
	funcMap := template.FuncMap{
		"join_string_array_sorted": func(input []string) string {
			sort.Strings(input)
			return strings.Join(input, ",")
		},
	}

	vars := make(map[string]interface{})
	vars["endpoints"] = endpoints

	base := template.Must(template.New("").Funcs(funcMap).Parse(`
{{ $tick := "` + "```" + `" }}
# Api

| Method | Uri | Description | Status Codes |
| --- | --- | --- | --- |
{{- range .endpoints }}
| {{.Method}} | {{.Path}} | {{.Description}} | {{join_string_array_sorted .StatusCodes}} |
{{- end }}

# Auto generated via
{{ $tick }}
make generate-docs-api-overview
{{ $tick }}
`))
	var tpl bytes.Buffer
	base.Execute(&tpl, vars)

	return strings.TrimSpace(tpl.String())
}

var apiOverviewCMD = &cobra.Command{
	Use:   "api-overview",
	Short: "Create markdown for api contents via openapi",
	Run: func(cmd *cobra.Command, args []string) {
		builder := DocsApiOverview{}
		b, _ := ioutil.ReadAll(os.Stdin)
		endpoints := builder.GetEndpoints(b)
		fmt.Println(builder.Render(endpoints))
	},
}

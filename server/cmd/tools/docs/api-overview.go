package docs

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

type Endpoint struct {
	Path        string   `json:"path"`
	Method      string   `json:"method"`
	Description string   `json:"description"`
	StatusCodes []string `json:"status_codes"`
	Tag         string   `json:"tag"`
}

type DocsApiOverview struct{}

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
		b, _ := ioutil.ReadAll(os.Stdin)
		s, _ := openapi3.NewSwaggerLoader().LoadSwaggerFromData(b)

		var endpoints []Endpoint
		for k, path := range s.Paths {
			for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete} {
				operation := path.GetOperation(method)
				if operation == nil {
					continue
				}

				statusCodes := make([]string, 0, len(operation.Responses))
				for k := range operation.Responses {
					statusCodes = append(statusCodes, k)
				}

				sort.Strings(statusCodes)
				tag := ""
				if len(operation.Tags) > 0 {
					tag = operation.Tags[0]
				}

				endpoint := Endpoint{
					Path:        k,
					Method:      method,
					Description: operation.Description,
					StatusCodes: statusCodes,
					Tag:         tag,
				}

				endpoints = append(endpoints, endpoint)
			}
		}

		sort.Slice(endpoints, func(i, j int) bool {
			return endpoints[i].Tag < endpoints[j].Tag
		})

		builder := DocsApiOverview{}
		fmt.Println(builder.Render(endpoints))
	},
}

package experimental

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type LinkServiceToOpenApiPathTag struct {
	Service string
	PathTag string
}

type DocEvent struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	EventType   string `json:"event_type"`
	From        string `json:"from"`
	OpenPathTag string `json:"openapi_path_tag,omitempty"`
}

func mainExtractEvents(cmd *cobra.Command, args []string) {
	argServices, _ := cmd.Flags().GetBool("services")
	argEvents, _ := cmd.Flags().GetBool("events")
	argEnrich, _ := cmd.Flags().GetBool("enrich")

	root, _ := cmd.Flags().GetString("root")
	if root == "" {
		fmt.Println("Root cant be empty --root")
		os.Exit(1)
	}

	dirs := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil

		}

		dirs = append(dirs, path)
		return nil
	})

	if err != nil {
		panic(err)
	}

	if argServices {
		services := make([]LinkServiceToOpenApiPathTag, 0)
		for _, lookIn := range dirs {
			fset := token.NewFileSet() // positions are relative to fset
			d, err := parser.ParseDir(fset, lookIn, nil, parser.ParseComments)
			if err != nil {
				fmt.Println(err)
				return
			}

			services = append(services, findServices(fset, d)...)
		}
		b, _ := json.Marshal(services)
		fmt.Println(string(b))
		os.Exit(0)
	}

	if argEvents {
		events := make([]DocEvent, 0)
		for _, lookIn := range dirs {
			fset := token.NewFileSet() // positions are relative to fset
			d, err := parser.ParseDir(fset, lookIn, nil, parser.ParseComments)
			if err != nil {
				fmt.Println(err)
				return
			}

			events = append(events, findEvents(fset, d)...)
		}
		b, _ := json.Marshal(events)
		fmt.Println(string(b))
		os.Exit(0)
	}

	if argEnrich {
		services := make([]LinkServiceToOpenApiPathTag, 0)
		for _, lookIn := range dirs {
			fset := token.NewFileSet() // positions are relative to fset
			d, err := parser.ParseDir(fset, lookIn, nil, parser.ParseComments)
			if err != nil {
				fmt.Println(err)
				return
			}

			services = append(services, findServices(fset, d)...)
		}

		events := make([]DocEvent, 0)
		for _, lookIn := range dirs {
			fset := token.NewFileSet() // positions are relative to fset
			d, err := parser.ParseDir(fset, lookIn, nil, parser.ParseComments)
			if err != nil {
				fmt.Println(err)
				return
			}

			events = append(events, findEvents(fset, d)...)
		}

		for i, event := range events {
			for _, service := range services {
				if event.Name == service.Service {
					event.OpenPathTag = service.PathTag
					events[i] = event
				}
			}
		}

		b, _ := json.Marshal(events)
		fmt.Println(string(b))

		os.Exit(0)
	}
}

func findServices(fset *token.FileSet, d map[string]*ast.Package) []LinkServiceToOpenApiPathTag {
	links := make([]LinkServiceToOpenApiPathTag, 0)
	for _, f := range d {
		for _, f := range f.Files {
			link := findServiceLinkedToOpenapi(fset, f)
			if link == (LinkServiceToOpenApiPathTag{}) {
				continue
			}
			links = append(links, link)
		}
	}
	return links
}

func findEvents(fset *token.FileSet, d map[string]*ast.Package) []DocEvent {
	found := make([]DocEvent, 0)
	for _, f := range d {
		for _, f := range f.Files {

			events := findEventsListening(fset, f)
			if len(events) > 0 {
				found = append(found, events...)
			}

			events = findEventsEmits(fset, f)
			if len(events) >= 0 {
				found = append(found, events...)
			}
		}
	}
	return found
}

func findServiceLinkedToOpenapi(fset *token.FileSet, f *ast.File) LinkServiceToOpenApiPathTag {
	link := LinkServiceToOpenApiPathTag{}

	for _, fc := range f.Decls {
		fn, ok := fc.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fn.Doc == nil {
			return link
		}

		if fn.Name.String() != "NewService" {
			continue
		}

		lookFor := "@openapi.path.tag: "
		for _, comment := range fn.Doc.List {
			if strings.Contains(comment.Text, lookFor) {
				parts := strings.Split(comment.Text, lookFor)
				var typeNameBuf bytes.Buffer
				err := printer.Fprint(&typeNameBuf, fset, fn.Type.Results.List[0].Type)
				if err != nil {
					log.Fatalf("failed printing %s", err)
				}

				structName := typeNameBuf.String()
				link.Service = structName
				link.PathTag = parts[1]
				return link
			}
		}
	}
	return link
}

func findEventsListening(fset *token.FileSet, f *ast.File) []DocEvent {
	data := make([]DocEvent, 0)

	for _, fc := range f.Decls {
		fn, ok := fc.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Doc == nil {
			continue
		}

		lookFor := "@event.listen: "
		for _, comment := range fn.Doc.List {
			if strings.Contains(comment.Text, lookFor) {

				var typeNameBuf bytes.Buffer
				err := printer.Fprint(&typeNameBuf, fset, fn.Recv.List[0].Type)
				if err != nil {
					log.Fatalf("failed printing %s", err)
				}
				structName := typeNameBuf.String()

				parts := strings.Split(comment.Text, lookFor)
				kind := parts[1]
				kind = strings.TrimSpace(kind)
				data = append(data, DocEvent{
					Name:      structName,
					Kind:      kind,
					EventType: "listen",
					From:      fn.Name.String(),
				})
			}
		}
	}
	return data
}

func findEventsEmits(fset *token.FileSet, f *ast.File) []DocEvent {
	data := make([]DocEvent, 0)

	for _, fc := range f.Decls {
		fn, ok := fc.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Doc == nil {
			continue
		}

		lookFor := "@event.emit: "
		for _, comment := range fn.Doc.List {
			if strings.Contains(comment.Text, lookFor) {
				var typeNameBuf bytes.Buffer
				err := printer.Fprint(&typeNameBuf, fset, fn.Recv.List[0].Type)
				if err != nil {
					log.Fatalf("failed printing %s", err)
				}
				structName := typeNameBuf.String()

				parts := strings.Split(comment.Text, lookFor)
				kind := parts[1]
				kind = strings.TrimSpace(kind)
				data = append(data, DocEvent{
					Name:      structName,
					Kind:      kind,
					EventType: "emit",
					From:      fn.Name.String(),
				})
			}
		}
	}
	return data
}

var extractEventsCMD = &cobra.Command{
	Use:   "extract-events",
	Short: "Extract events from the source code based on annotations",
	Long: `
Wondered how hard it would be to use the code to help document the flow of events.
It doesnt work well yet, not sure if its my lack of knowledge with d3,
I am missing some metadata, I will return to this


We look for the following annotations:
- @event.emit
- @event.listen
- @openapi.path.tag

go run main.go --config=../config/dev.config.yaml tools experimental extract-events --enrich
	`,
	Run: mainExtractEvents,
}

func init() {
	extractEventsCMD.Flags().Bool("services", false, "Find link from services to openapi paths")
	extractEventsCMD.Flags().Bool("events", false, "Find code that emits or listens to events")
	extractEventsCMD.Flags().Bool("enrich", false, "Find code that emits or listens to events")
	extractEventsCMD.Flags().String("root", "", "Source code to look for annotations")
}

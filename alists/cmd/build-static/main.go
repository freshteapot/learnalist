package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	staticSiteFolder := flag.String("static", "", "path to static site builder")
	flag.Parse()

	parts := strings.Split("--cleanDestinationDir -e alist --config=config/alist/config.toml", " ")
	cmd := exec.Command("hugo", parts...)
	cmd.Dir = *staticSiteFolder
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))
}

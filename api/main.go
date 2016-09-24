package main

import (
	"flag"

	"github.com/freshteapot/learnalist/api/api"
)

func main() {
	database := flag.String("database", "/tmp/api.db", "The database.")
	domain := flag.String("domain", "learnalist.net", "The domain.")
	basicAuth := flag.String("basicAuth", "", "Single user with basic auth username:password.")
	port := flag.Int("port", 80, "Port to listen on.")
	flag.Parse()

	api.SetDomain(*domain)
	if *basicAuth != "" {
		api.UseBasicAuth(*basicAuth)
	}

	api.Run(*port, *database)
}

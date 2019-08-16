package main

import (
	"flag"
	"log"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/server"
)

func main() {
	databaseName := flag.String("database", "/tmp/api.db", "The database.")
	domain := flag.String("domain", "learnalist.net", "The domain.")
	port := flag.Int("port", 80, "Port to listen on.")
	corsAllowedOrigins := flag.String("cors-allowed-origins", "", "Use , between allowed domains.")
	staticSiteFolder := flag.String("static", "", "path to static site builder")
	flag.Parse()

	*staticSiteFolder = strings.TrimRight(*staticSiteFolder, "/")

	if *staticSiteFolder == "" {
		log.Fatal("Will need the path to static site, add -static=XXX")
	}

	serverConfig := server.Config{
		Port:             *port,
		Domain:           *domain,
		CorsAllowOrigins: *corsAllowedOrigins,
		StaticSiteFolder: *staticSiteFolder,
	}
	server.Init(serverConfig)

	db := database.NewDB(*databaseName)
	// Setup access control layer.
	acl := acl.NewAclFromModel(*databaseName)

	server.InitApi(db, acl)
	server.InitAlists(acl)

	server.Run()
}

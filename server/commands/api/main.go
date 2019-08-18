package main

import (
	"flag"
	"log"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/server"
)

func main() {
	databaseName := flag.String("database", "/tmp/api.db", "The database.")
	domain := flag.String("domain", "learnalist.net", "The domain.")
	port := flag.Int("port", 80, "Port to listen on.")
	corsAllowedOrigins := flag.String("cors-allowed-origins", "", "Use , between allowed domains.")
	staticSiteFolder := flag.String("site-builder-dir", "", "path to static site builder")
	siteCacheFolder := flag.String("site-cache-dir", "", "path to site cache")
	flag.Parse()

	*staticSiteFolder = strings.TrimRight(*staticSiteFolder, "/")
	*siteCacheFolder = strings.TrimRight(*siteCacheFolder, "/")

	if *staticSiteFolder == "" {
		log.Fatal("Will need the path to site builder directory, add -site-builder-dir=XXX")
	}

	if *siteCacheFolder == "" {
		log.Fatal("Will need the path to site cache directory, add -site-cache-dir=XXX")
	}

	serverConfig := server.Config{
		Port:             *port,
		Domain:           *domain,
		CorsAllowOrigins: *corsAllowedOrigins,
		StaticSiteFolder: *staticSiteFolder,
		SiteCacheFolder:  *siteCacheFolder,
	}
	server.Init(serverConfig)

	cron := cron.NewCron()
	db := database.NewDB(*databaseName)
	// Setup access control layer.
	acl := acl.NewAclFromModel(*databaseName)
	server.InitApi(db, acl)
	server.InitAlists(cron, acl)

	server.Run()
}

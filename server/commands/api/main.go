package main

import (
	"flag"
	"log"
	"strings"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
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

	siteCacheFolder := flag.String("site-cache-dir", "", "path to site cache")
	hugoFolder := flag.String("hugo-dir", "", "path to static site builder")
	flag.Parse()

	*hugoFolder = strings.TrimRight(*hugoFolder, "/")
	*siteCacheFolder = strings.TrimRight(*siteCacheFolder, "/")

	if *hugoFolder == "" {
		log.Fatal("Will need the path to site builder directory, add -hugo-dir=XXX")
	}

	if *siteCacheFolder == "" {
		log.Fatal("Will need the path to site cache directory, add -site-cache-dir=XXX")
	}

	serverConfig := server.Config{
		Port:             *port,
		Domain:           *domain,
		CorsAllowOrigins: *corsAllowedOrigins,
		HugoFolder:       *hugoFolder,
		SiteCacheFolder:  *siteCacheFolder,
	}
	server.Init(serverConfig)

	masterCron := cron.NewCron()
	db := database.NewDB(*databaseName)
	hugoHelper := hugo.NewHugoHelper(serverConfig.HugoFolder, masterCron)
	hugoHelper.RegisterCronJob()

	// Setup access control layer.
	acl := acl.NewAclFromModel(*databaseName)
	server.InitApi(db, acl, hugoHelper)
	server.InitAlists(acl, hugoHelper)

	server.Run()
}

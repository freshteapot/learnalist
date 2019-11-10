package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	aclSqlite "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/freshteapot/learnalist-api/server/server"
)

func main() {
	databaseName := flag.String("database", "/tmp/api.db", "The database.")
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

	if !utils.IsDir(*hugoFolder) {
		log.Fatal(fmt.Sprintf("%s is not a directory", *hugoFolder))
	}

	if *siteCacheFolder == "" {
		log.Fatal("Will need the path to site cache directory, add -site-cache-dir=XXX")
	}

	if !utils.IsDir(*siteCacheFolder) {
		log.Fatal(fmt.Sprintf("%s is not a directory", *siteCacheFolder))
	}
	// Convert paths to absolute, allowing /../x
	*hugoFolder, _ = filepath.Abs(*hugoFolder)
	*siteCacheFolder, _ = filepath.Abs(*siteCacheFolder)

	serverConfig := server.Config{
		Port:             *port,
		CorsAllowOrigins: *corsAllowedOrigins,
		HugoFolder:       *hugoFolder,
		SiteCacheFolder:  *siteCacheFolder,
	}
	server.Init(serverConfig)

	masterCron := cron.NewCron()

	// *databaseName = "root:mysecretpassword@/learnalistapi"
	db := database.NewDB(*databaseName)
	hugoHelper := hugo.NewHugoHelper(serverConfig.HugoFolder, masterCron, serverConfig.SiteCacheFolder)
	hugoHelper.RegisterCronJob()

	// Setup access control layer.
	acl := aclSqlite.NewAcl(db)
	dal := models.NewDAL(db, acl)
	server.InitApi(db, acl, dal, hugoHelper)
	server.InitAlists(acl, dal, hugoHelper)

	server.Run()
}

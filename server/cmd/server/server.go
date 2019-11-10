package server

import (
	"fmt"

	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	aclSqlite "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/freshteapot/learnalist-api/server/server"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the server {api,backend}",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start the server")
		databaseName := viper.GetString("server.sqlite.database")
		port := viper.GetString("server.port")
		corsAllowedOrigins := viper.GetString("server.cors.allowedOrigins")
		siteCacheFolder := viper.GetString("server.siteCacheDirectory") // "path to site cache"
		hugoFolder := viper.GetString("server.hugoDirectory")           // "path to static site builder

		hugoFolder = strings.TrimRight(hugoFolder, "/")
		siteCacheFolder = strings.TrimRight(siteCacheFolder, "/")

		// Convert paths to absolute, allowing /../x
		hugoFolder, _ = filepath.Abs(hugoFolder)
		siteCacheFolder, _ = filepath.Abs(siteCacheFolder)

		if hugoFolder == "" {
			log.Fatal("You might have forgotten to set the path to hugo directory: server.hugoDirectory")
		}

		if !utils.IsDir(hugoFolder) {
			log.Fatal(fmt.Sprintf("%s is not a directory", hugoFolder))
		}

		if siteCacheFolder == "" {
			log.Fatal("You might have forgotten to set the path to site cache directory: server.siteCacheDirectory")
		}

		if !utils.IsDir(siteCacheFolder) {
			log.Fatal(fmt.Sprintf("%s is not a directory", siteCacheFolder))
		}

		serverConfig := server.Config{
			Port:             port,
			CorsAllowOrigins: corsAllowedOrigins,
			HugoFolder:       hugoFolder,
			SiteCacheFolder:  siteCacheFolder,
		}
		server.Init(serverConfig)

		masterCron := cron.NewCron()

		// databaseName = "root:mysecretpassword@/learnalistapi"
		db := database.NewDB(databaseName)
		hugoHelper := hugo.NewHugoHelper(serverConfig.HugoFolder, masterCron, serverConfig.SiteCacheFolder)
		hugoHelper.RegisterCronJob()

		// Setup access control layer.
		acl := aclSqlite.NewAcl(db)
		dal := models.NewDAL(db, acl)
		server.InitApi(db, acl, dal, hugoHelper)
		server.InitAlists(acl, dal, hugoHelper)

		server.Run()
	},
}

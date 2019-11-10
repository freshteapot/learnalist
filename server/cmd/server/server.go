package server

import (
	"fmt"
	"os"

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
		databaseName := viper.GetString("server.sqlite.database")
		port := viper.GetString("server.port")
		corsAllowedOrigins := viper.GetString("server.cors.allowedOrigins")

		// "path to static site builder
		hugoFolder, err := utils.CmdParsePathToFolder("server.hugoDirectory", viper.GetString("server.hugoDirectory"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// "path to site cache"
		siteCacheFolder, err := utils.CmdParsePathToFolder("server.siteCacheDirectory", viper.GetString("server.siteCacheDirectory"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
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

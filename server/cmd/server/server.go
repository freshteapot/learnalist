package server

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/freshteapot/learnalist-api/server/server"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the server {api,backend}",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		googleOauthConfig := oauth.NewGoogle(oauth.GoogleConfig{
			Key:    viper.GetString("server.loginWith.google.clientID"),
			Secret: viper.GetString("server.loginWith.google.clientSecret"),
			Server: viper.GetString("server.loginWith.google.server"),
		})
		viper.Set("server.loginWith.google.clientSecret", "***")

		oauthHandlers := &oauth.Handlers{
			Google: googleOauthConfig,
		}

		databaseName := viper.GetString("server.sqlite.database")
		port := viper.GetString("server.port")
		corsAllowedOrigins := viper.GetString("server.cors.allowedOrigins")

		// "path to static site builder
		hugoFolder, err := utils.CmdParsePathToFolder("server.hugo.directory", viper.GetString("server.hugo.directory"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		hugoEnvironment := viper.GetString("server.hugo.environment")
		if hugoEnvironment == "" {
			fmt.Println("server.hugo.environment is missing")
			os.Exit(1)
		}

		hugoExternal := viper.GetBool("server.hugo.external")
		if hugoEnvironment == "" {
			fmt.Println("server.hugo.external is missing")
			os.Exit(1)
		}

		// "path to site cache"
		siteCacheFolder, err := utils.CmdParsePathToFolder("server.siteCacheDirectory", viper.GetString("server.siteCacheDirectory"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		// A hack would be to access it via
		loginCookie := authenticate.CookieConfig{
			Domain: viper.GetString("server.cookie.domain"),
			Secure: viper.GetBool("server.cookie.secure"),
		}

		logger.WithFields(logrus.Fields{
			"settings": viper.AllSettings(),
		}).Info("server startup")

		serverConfig := server.Config{
			Port:             port,
			CorsAllowOrigins: corsAllowedOrigins,
			HugoFolder:       hugoFolder,
			SiteCacheFolder:  siteCacheFolder,
		}
		server.Init(serverConfig)

		authenticate.SetLoginCookieConfig(loginCookie)

		masterCron := cron.NewCron()

		// databaseName = "root:mysecretpassword@/learnalistapi"
		db := database.NewDB(databaseName)
		hugoHelper := hugo.NewHugoHelper(serverConfig.HugoFolder, hugoEnvironment, hugoExternal, masterCron, serverConfig.SiteCacheFolder)
		hugoHelper.RegisterCronJob()

		// Setup access control layer.
		acl := aclStorage.NewAcl(db)
		userSession := userStorage.NewUserSession(db)
		userFromIDP := userStorage.NewUserFromIDP(db)
		userWithUsernameAndPassword := userStorage.NewUserWithUsernameAndPassword(db)
		oauthHandler := oauthStorage.NewOAuthReadWriter(db)
		dal := models.NewDAL(db, acl, userSession, userFromIDP, userWithUsernameAndPassword, oauthHandler)

		server.InitApi(db, acl, dal, hugoHelper, oauthHandlers)
		server.InitAlists(acl, dal, hugoHelper)
		server.Run()
	},
}

func init() {
	viper.BindEnv("server.loginWith.google.clientID", "LOGIN_WITH_GOOGLE_ID")
	viper.BindEnv("server.loginWith.google.clientSecret", "LOGIN_WITH_GOOGLE_SECRET")
	viper.BindEnv("server.loginWith.google.server", "LOGIN_WITH_GOOGLE_SERVER")
	viper.BindEnv("server.hugo.external", "HUGO_EXTERNAL")
}

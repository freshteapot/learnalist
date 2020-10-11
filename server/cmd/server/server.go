package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	alistStorage "github.com/freshteapot/learnalist-api/server/api/alist/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/api"
	"github.com/freshteapot/learnalist-api/server/api/database"
	labelStorage "github.com/freshteapot/learnalist-api/server/api/label/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/models"
	apiUserStorage "github.com/freshteapot/learnalist-api/server/api/user/sqlite"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/assets"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/freshteapot/learnalist-api/server/server"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the server {api,backend}",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()

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
		// Assets
		assetsDirectory := viper.GetString("server.assets.directory")
		// Static site
		hugoFolder, err := utils.CmdParsePathToFolder("hugo.directory", viper.GetString("hugo.directory"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		hugoEnvironment := viper.GetString("hugo.environment")
		if hugoEnvironment == "" {
			fmt.Println("hugo.environment is missing")
			os.Exit(1)
		}

		hugoExternal := viper.GetBool("hugo.external")

		// A hack would be to access it via
		loginCookie := authenticate.CookieConfig{
			Domain: viper.GetString("server.cookie.domain"),
			Secure: viper.GetBool("server.cookie.secure"),
		}

		logger.WithFields(logrus.Fields{
			"settings": viper.AllSettings(),
		}).Info("server startup")

		// This now works for the "application"
		eventsVia := viper.GetString("server.events.via")
		if eventsVia == "memory" {
			event.SetBus(event.NewMemoryBus())
		}

		if eventsVia == "nats" {
			natsServer := viper.GetString("server.events.nats.server")
			stanClusterID := viper.GetString("server.events.stan.clusterID")
			stanClientID := viper.GetString("server.events.stan.clientID")
			opts := []nats.Option{nats.Name("lal-go-server")}
			nc, err := nats.Connect(natsServer, opts...)

			if err != nil {
				panic(err)
			}
			defer nc.Close()

			bus := event.NewNatsBus(stanClusterID, stanClientID, nc, logger)
			bus.Subscribe(event.TopicMonolog, func() {
			})
			event.SetBus(bus)
		}
		event.GetBus().Start()

		serverConfig := server.Config{
			Port:             port,
			CorsAllowOrigins: corsAllowedOrigins,
		}
		server.Init(serverConfig)

		authenticate.SetLoginCookieConfig(loginCookie)

		masterCron := cron.NewCron()
		server.SetCron(masterCron)

		// databaseName = "root:mysecretpassword@/learnalistapi"
		db := database.NewDB(databaseName)
		hugoHelper := hugo.NewHugoHelper(hugoFolder, hugoEnvironment, hugoExternal, masterCron, logger)
		hugoHelper.RegisterCronJob()

		// Setup access control layer.
		acl := aclStorage.NewAcl(db)
		userSession := userStorage.NewUserSession(db)
		userFromIDP := userStorage.NewUserFromIDP(db)
		userWithUsernameAndPassword := userStorage.NewUserWithUsernameAndPassword(db)
		oauthHandler := oauthStorage.NewOAuthReadWriter(db)
		labels := labelStorage.NewLabel(db)
		storageAlist := alistStorage.NewAlist(db, logger)
		storageApiUser := apiUserStorage.NewUser(db)
		dal := models.NewDAL(
			acl,
			storageApiUser,
			storageAlist,
			labels, userSession, userFromIDP, userWithUsernameAndPassword, oauthHandler)

		userManagement := user.NewManagement(
			userStorage.NewSqliteManagementStorage(db),
			hugoHelper,
			event.NewInsights(logger),
		)

		apiManager := api.NewManager(dal, userManagement, acl, "", hugoHelper, *oauthHandlers, logger)

		// TODO how to hook up sse https://gist.github.com/freshteapot/d467adb7cb082d2d056205deb38a9694
		spacedRepetitionService := spaced_repetition.NewService(db)
		assetService := assets.NewService(assetsDirectory, acl, assets.NewSqliteRepository(db), logger.WithField("context", "assets-service"))
		assetService.InitCheck()

		server.InitApi(apiManager, assetService, spacedRepetitionService)
		server.InitAlists(acl, dal, hugoHelper)

		go func() {
			server.Run()
		}()

		ctx, cancel := context.WithCancel(context.Background())
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
		case <-sigterm:
			log.Println("terminating: via signal")
		}

		event.GetBus().Close()
		cancel()
		// TODO I dont think this is needed
		//time.Sleep(2 * time.Second)
		fmt.Println("done")
	},
}

func init() {
	viper.BindEnv("server.loginWith.google.clientID", "LOGIN_WITH_GOOGLE_ID")
	viper.BindEnv("server.loginWith.google.clientSecret", "LOGIN_WITH_GOOGLE_SECRET")
	viper.BindEnv("server.loginWith.google.server", "LOGIN_WITH_GOOGLE_SERVER")

	// If the events are not complicated, then this should work for memory or nats
	viper.SetDefault("server.events.via", "memory")
	viper.BindEnv("server.events.via", "EVENTS_VIA")

	viper.BindEnv("hugo.external", "HUGO_EXTERNAL")
}

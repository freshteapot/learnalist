package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	serviceAlists "github.com/freshteapot/learnalist-api/server/alists/server"
	alistStorage "github.com/freshteapot/learnalist-api/server/api/alist/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/api"
	"github.com/freshteapot/learnalist-api/server/api/database"
	labelStorage "github.com/freshteapot/learnalist-api/server/api/label/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/app_settings"
	"github.com/freshteapot/learnalist-api/server/pkg/assets"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/plank"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/staticsite/hugo"

	oauthApi "github.com/freshteapot/learnalist-api/server/pkg/oauth/api"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userApi "github.com/freshteapot/learnalist-api/server/pkg/user/api"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"

	eventHugoGlue "github.com/freshteapot/learnalist-api/server/pkg/event/hugo"
	"github.com/freshteapot/learnalist-api/server/pkg/event/staticsite"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/freshteapot/learnalist-api/server/server"
)

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the server {api,backend}",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()

		viper.SetDefault("server.userRegisterKey", "")
		viper.BindEnv("server.userRegisterKey", "USER_REGISTER_KEY")

		googleOauthConfig := oauth.NewGoogle(oauth.GoogleConfig{
			Key:       viper.GetString("server.loginWith.google.clientID"),
			Secret:    viper.GetString("server.loginWith.google.clientSecret"),
			Server:    viper.GetString("server.loginWith.google.server"),
			Audiences: viper.GetStringSlice("server.loginWith.google.audiences"),
		})
		viper.Set("server.loginWith.google.clientSecret", "***")

		var appleWebAudience oauth.AppleConfig
		viper.UnmarshalKey("server.loginWith.appleid.web", &appleWebAudience)

		var appleAudiences []oauth.AppleConfig
		viper.UnmarshalKey("server.loginWith.appleid.apps", &appleAudiences)
		appleAudiences = append(appleAudiences, appleWebAudience)

		appleIDOauthConfig := oauth.NewAppleID(appleWebAudience, appleAudiences)

		// Hiding cert from the allsettings
		hideCertAppleAudiences := appleAudiences
		for index := range hideCertAppleAudiences {
			hideCertAppleAudiences[index].Cert = "***"
		}
		viper.Set("server.loginWith.appleid.web.cert", "***")
		viper.Set("server.loginWith.appleid.apps", hideCertAppleAudiences)

		oauthHandlers := oauth.NewHandlers()
		oauthHandlers.AddGoogle(googleOauthConfig)
		oauthHandlers.AddAppleID(appleIDOauthConfig)

		userRegisterKey := viper.GetString("server.userRegisterKey")
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

		// A hack would be to access it via
		loginCookie := authenticate.CookieConfig{
			Domain: viper.GetString("server.cookie.domain"),
			Secure: viper.GetBool("server.cookie.secure"),
		}

		logger.WithFields(logrus.Fields{
			"settings": viper.AllSettings(),
		}).Info("server startup")

		event.SetupEventBus(logger.WithField("context", "event-bus-setup"))

		if !viper.GetBool("staticsite.external") {
			hugoEnvironment := viper.GetString("hugo.environment")
			if hugoEnvironment == "" {
				fmt.Println("hugo.environment is missing")
				os.Exit(1)
			}

			hugoHelper := hugo.NewHugoHelper(hugoFolder, hugoEnvironment, logger.WithField("context", "static-site-hugo"))
			hugoHelper.ListenForEvents()
		}

		serverConfig := server.Config{
			Port:             port,
			CorsAllowOrigins: corsAllowedOrigins,
		}
		server.Init(serverConfig)

		authenticate.SetLoginCookieConfig(loginCookie)

		db := database.NewDB(databaseName)
		// Slightly decoupled, but life goes on
		publishDirectory := fmt.Sprintf(hugo.RealtivePathPublic, hugoFolder)

		// Setup access control layer.
		aclRepo := aclStorage.NewAcl(db)
		userSession := userStorage.NewUserSession(db)
		userFromIDP := userStorage.NewUserFromIDP(db)
		userWithUsernameAndPassword := userStorage.NewUserWithUsernameAndPassword(db)
		oauthHandler := oauthStorage.NewOAuthReadWriter(db)
		labels := labelStorage.NewLabel(db)
		storageAlist := alistStorage.NewAlist(db, logger)
		userInfoRepo := user.NewUserInfo(userStorage.NewUserInfo(db))

		glueHugo := eventHugoGlue.NewGlue(
			storageAlist,
			logger.WithField("context", "event-hugo-glue"))
		glueHugo.ListenForEvents()

		dal := models.NewDAL(
			aclRepo,
			storageAlist,
			labels,
			oauthHandler)

		userStorageRepo := userStorage.NewSqliteManagementStorage(db)

		userManagement := user.NewManagement(
			userStorageRepo,
			staticsite.NewSiteManagementViaEvents(),
			event.NewInsights(logger),
		)

		oauthApiService := oauthApi.NewService(
			userManagement,
			*oauthHandlers,
			userSession,
			userFromIDP,
			logger.WithField("context", "oauth-service"),
		)

		apiManager := api.NewManager(
			dal,
			userManagement,
			aclRepo,
			"",
			logger)

		// TODO how to hook up sse https://gist.github.com/freshteapot/d467adb7cb082d2d056205deb38a9694
		spacedRepetitionRepo := spaced_repetition.NewSqliteRepository(db)
		spacedRepetitionService := spaced_repetition.NewService(
			spacedRepetitionRepo,
			logger.WithField("context", "spacedRepetition-service"),
		)

		plankService := plank.NewService(
			plank.NewSqliteRepository(db),
			logger.WithField("context", "plank-service"),
		)

		assetService := assets.NewService(
			assetsDirectory,
			aclRepo,
			assets.NewSqliteRepository(db),
			logger.WithField("context", "assets-service"),
		)
		assetService.InitCheck()

		userService := userApi.NewService(
			*oauthHandlers,
			aclRepo,
			userManagement,
			userFromIDP,
			userSession,
			userWithUsernameAndPassword,
			userInfoRepo,
			userRegisterKey,
			logger.WithField("context", "user-service"))

		challengeRepo := challenge.NewSqliteRepository(db)
		challengeNotificationRepo := challengeRepo.(challenge.ChallengePushNotificationRepository)
		challengeService := challenge.NewService(
			challengeRepo,
			challengeNotificationRepo,
			aclRepo,
			logger.WithField("context", "challenge-service"),
		)

		mobileService := mobile.NewService(
			mobile.NewSqliteRepository(db),
			logger.WithField("context", "mobile-service"))

		remindService := remind.NewService(
			userInfoRepo,
			logger.WithField("context", "remind-service"))

		appSettingsService := app_settings.NewService(
			userInfoRepo,
			logger.WithField("context", "appSettings-service"),
		)

		dripfeedService := dripfeed.NewService(
			dripfeed.NewSqliteRepository(db),
			aclRepo,
			storageAlist,
			logger.WithField("context", "dripfeed-service"),
		)

		_ = acl.NewService(
			aclRepo,
			logger.WithField("context", "acl-service"),
		)

		server.InitApi(
			apiManager,
			userSession,
			userWithUsernameAndPassword,
			userService,
			assetService,
			spacedRepetitionService,
			plankService,
			challengeService,
			mobileService,
			remindService,
			appSettingsService,
			dripfeedService,
			oauthApiService,
		)

		server.InitAlists(
			serviceAlists.NewManager(
				aclRepo,
				dal,
				userSession,
				publishDirectory),
			userSession,
			userWithUsernameAndPassword,
			publishDirectory)

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
	},
}

func init() {
	viper.BindEnv("server.loginWith.google.clientID", "LOGIN_WITH_GOOGLE_ID")
	viper.BindEnv("server.loginWith.google.clientSecret", "LOGIN_WITH_GOOGLE_SECRET")
	viper.BindEnv("server.loginWith.google.server", "LOGIN_WITH_GOOGLE_SERVER")

	// If the events are not complicated, then this should work for memory or nats
	viper.SetDefault("server.events.via", "memory")
	viper.BindEnv("server.events.via", "EVENTS_VIA")

	viper.SetDefault("staticsite.external", false)
	viper.BindEnv("staticsite.external", "STATIC_SITE_EXTERNAL")
}

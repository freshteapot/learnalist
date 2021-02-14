package server

import (
	"github.com/freshteapot/learnalist-api/server/api/api"
	authenticateApi "github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/app_settings"
	"github.com/freshteapot/learnalist-api/server/pkg/assets"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/plank"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userInfo "github.com/freshteapot/learnalist-api/server/pkg/user/info"
)

func InitApi(
	apiManager *api.Manager,
	userService user.UserService,
	assetService assets.AssetService,
	spacedRepetitionService spaced_repetition.SpacedRepetitionService,
	plankService plank.PlankService,
	challengeService challenge.ChallengeService,
	mobileService mobile.MobileService,
	remindService remind.RemindService,
	appSettingsService app_settings.AppSettingsService,
	dripfeedService dripfeed.DripfeedService,
	userInfoService userInfo.UserInfoService,
) {

	authConfig := authenticate.Config{
		LookupBasic:  apiManager.Datastore.UserWithUsernameAndPassword().Lookup,
		LookupBearer: apiManager.Datastore.UserSession().GetUserUUIDByToken,
		Skip:         authenticateApi.Skip,
	}

	assetAuthConfig := authenticate.Config{
		LookupBasic:  apiManager.Datastore.UserWithUsernameAndPassword().Lookup,
		LookupBearer: apiManager.Datastore.UserSession().GetUserUUIDByToken,
		Skip:         assets.SkipAuth,
	}

	v1 := server.Group("/api/v1")

	v1.Use(authenticate.Auth(authConfig))

	v1.GET("/version", apiManager.V1GetVersion)

	v1.POST("/user/register", apiManager.V1PostRegister)
	v1.GET("/user/info/:uuid", userInfoService.V1GetUserInfo)
	v1.PATCH("/user/info/:uuid", userInfoService.V1PatchUserInfo)
	v1.POST("/user/login/idp", userService.LoginViaIDP)
	v1.POST("/user/login", apiManager.V1PostLogin)
	v1.POST("/user/logout", apiManager.V1PostLogout)
	v1.DELETE("/user/:uuid", apiManager.V1DeleteUser)

	// Route => handler
	v1.GET("/", apiManager.V1GetRoot)
	v1.GET("/alist/:uuid", apiManager.V1GetListByUUID)
	v1.GET("/alist/by/me", apiManager.V1GetListsByMe)

	//e.POST("/alist/v1", m.V1PostAlist)
	//e.POST("/alist/v2", m.V1PostAlist)
	//e.POST("/alist/v3", m.V1PostAlist)
	//e.POST("/alist/v4", m.V1PostAlist)
	v1.POST("/alist", apiManager.V1SaveAlist)
	v1.PUT("/share/alist", apiManager.V1ShareAlist)
	v1.PUT("/share/readaccess", apiManager.V1ShareListReadAccess)
	v1.PUT("/alist/:uuid", apiManager.V1SaveAlist)
	v1.DELETE("/alist/:uuid", apiManager.V1RemoveAlist)
	// Labels
	v1.POST("/labels", apiManager.V1PostUserLabel)
	v1.GET("/labels/by/me", apiManager.V1GetUserLabels)
	v1.DELETE("/labels/:label", apiManager.V1RemoveUserLabel)

	// Assets
	server.GET("/assets/*", assetService.GetAsset, authenticate.Auth(assetAuthConfig))
	v1.POST("/assets/upload", assetService.Upload)
	v1.PUT("/assets/share", assetService.Share)
	v1.DELETE("/assets/:uuid", assetService.DeleteEntry)

	// Oauth
	v1.GET("/oauth/google/redirect", apiManager.V1OauthGoogleRedirect)
	v1.GET("/oauth/google/callback", apiManager.V1OauthGoogleCallback)

	// Spaced Repetition
	srs := server.Group("/api/v1/spaced-repetition")
	srs.Use(authenticate.Auth(authConfig))
	srs.GET("/next", spacedRepetitionService.GetNext)
	srs.GET("/all", spacedRepetitionService.GetAll)
	srs.DELETE("/:uuid", spacedRepetitionService.DeleteEntry)
	srs.POST("/", spacedRepetitionService.SaveEntry)
	srs.POST("/viewed", spacedRepetitionService.EntryViewed)

	// Dripfeed service
	srs.GET("/overtime/active/:alistUUID", dripfeedService.ListActive) // Might be shit, need to pick one and go with it
	srs.POST("/overtime", dripfeedService.Create)
	srs.DELETE("/overtime", dripfeedService.Delete)

	// Plank
	plank := server.Group("/api/v1/plank")
	plank.Use(authenticate.Auth(authConfig))
	plank.GET("/history", plankService.History)
	plank.DELETE("/:uuid", plankService.DeletePlankRecord)
	plank.POST("/", plankService.RecordPlank)

	// Challenge
	v1.GET("/challenges/:userUUID", challengeService.Challenges)
	challengeV1 := server.Group("/api/v1/challenge")
	challengeV1.Use(authenticate.Auth(authConfig))

	challengeV1.PUT("/:uuid/join", challengeService.Join)
	challengeV1.PUT("/:uuid/leave", challengeService.Leave)
	challengeV1.POST("/", challengeService.Create)
	challengeV1.GET("/:uuid", challengeService.Get)
	challengeV1.DELETE("/:uuid", challengeService.Delete)

	// Mobile
	mobileV1 := server.Group("/api/v1/mobile")
	mobileV1.Use(authenticate.Auth(authConfig))
	mobileV1.POST("/register-device", mobileService.RegisterDevice)

	// Remind Service
	remindV1 := server.Group("/api/v1/remind")
	remindV1.Use(authenticate.Auth(authConfig))
	remindV1.GET("/daily/:appIdentifier", remindService.GetDailySettings)
	remindV1.DELETE("/daily/:appIdentifier", remindService.DeleteDailySettings)
	remindV1.PUT("/daily/", remindService.SetDailySettings)

	// App Settings Service
	settingsV1 := server.Group("/api/v1/app-settings")
	settingsV1.Use(authenticate.Auth(authConfig))
	settingsV1.PUT("/remind_v1", appSettingsService.SaveRemindV1)
}

package server

import (
	"github.com/freshteapot/learnalist-api/server/api/api"
	authenticateApi "github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/assets"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
)

func InitApi(apiManager *api.Manager, assetService assets.AssetService, spacedRepetitionService spaced_repetition.SpacedRepetitionService) {

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
}

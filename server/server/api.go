package server

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/api"
	authenticateApi "github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/jmoiron/sqlx"

	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
)

func InitApi(db *sqlx.DB, acl acl.Acl, dal *models.DAL, hugoHelper *hugo.HugoHelper, oauthHandlers *oauth.Handlers) {

	m := api.Manager{
		Datastore:     dal,
		Acl:           acl,
		HugoHelper:    *hugoHelper,
		OauthHandlers: *oauthHandlers,
	}

	authConfig := authenticate.Config{
		LookupBasic:  m.Datastore.UserWithUsernameAndPassword().Lookup,
		LookupBearer: m.Datastore.UserSession().GetUserUUIDByToken,
		Skip:         authenticateApi.Skip,
	}

	v1 := server.Group("/api/v1")

	v1.Use(authenticate.Auth(authConfig))

	v1.GET("/version", m.V1GetVersion)

	v1.POST("/user/register", m.V1PostRegister)
	v1.POST("/user/login", m.V1PostLogin)
	v1.POST("/user/logout", m.V1PostLogout)

	// Route => handler
	v1.GET("/", m.V1GetRoot)
	v1.GET("/alist/:uuid", m.V1GetListByUUID)
	v1.GET("/alist/by/me", m.V1GetListsByMe)

	//e.POST("/alist/v1", m.V1PostAlist)
	//e.POST("/alist/v2", m.V1PostAlist)
	//e.POST("/alist/v3", m.V1PostAlist)
	//e.POST("/alist/v4", m.V1PostAlist)
	v1.POST("/alist", m.V1SaveAlist)
	v1.PUT("/share/alist", m.V1ShareAlist)
	v1.PUT("/share/readaccess", m.V1ShareListReadAccess)
	v1.PUT("/alist/:uuid", m.V1SaveAlist)
	v1.DELETE("/alist/:uuid", m.V1RemoveAlist)
	// Labels
	v1.POST("/labels", m.V1PostUserLabel)
	v1.GET("/labels/by/me", m.V1GetUserLabels)
	v1.DELETE("/labels/:label", m.V1RemoveUserLabel)

	// Oauth
	v1.GET("/oauth/google/redirect", m.V1OauthGoogleRedirect)
	v1.GET("/oauth/google/callback", m.V1OauthGoogleCallback)
}

package server

import (
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/api"
	"github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/jmoiron/sqlx"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func InitApi(db *sqlx.DB, acl acl.Acl, dal *models.DAL, hugoHelper *hugo.HugoHelper) {
	m := api.Manager{
		Datastore:  dal,
		Acl:        acl,
		HugoHelper: *hugoHelper,
	}

	authenticate.LookUp = m.Datastore.GetUserByCredentials
	v1 := server.Group("/api/v1")
	if config.CorsAllowOrigins != "" {
		allowOrigins := strings.Split(config.CorsAllowOrigins, ",")
		v1.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: allowOrigins,
			AllowMethods: []string{http.MethodOptions, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
			AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderOrigin, echo.HeaderContentType},
		}))
	}

	v1.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper:   authenticate.SkipBasicAuth,
		Validator: authenticate.ValidateBasicAuth,
	}))

	v1.GET("/version", m.V1GetVersion)

	v1.POST("/register", m.V1PostRegister)

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
	v1.DELETE("/labels/:uuid", m.V1RemoveUserLabel)
}

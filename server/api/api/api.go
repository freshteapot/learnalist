package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Env exposing the data abstraction layer
type Env struct {
	Datastore        models.Datastore
	Port             int
	DatabaseName     string
	CorsAllowOrigins string
	Acl              acl.Acl
}

type HttpResponseMessage struct {
	Message string `json:"message"`
}

// @todo change this to be a configure file
var domain string

// SetDomain set the domain this api is associated with.
func SetDomain(_domain string) {
	domain = _domain
}

// Run This starts the api listening on the port supplied
func Run(env Env) {
	db := database.NewDB(env.DatabaseName)
	// Setup access control layer.
	acl := acl.NewAclFromModel(env.DatabaseName)
	env.Datastore = models.NewDAL(db, acl)

	// Echo instance
	e := echo.New()
	e.HideBanner = true
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 9,
	}))

	authenticate.LookUp = env.Datastore.GetUserByCredentials

	v1 := e.Group("/v1")
	if env.CorsAllowOrigins != "" {
		allowOrigins := strings.Split(env.CorsAllowOrigins, ",")
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: allowOrigins,
			AllowMethods: []string{http.MethodOptions, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
			AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderOrigin, echo.HeaderContentType},
		}))
	}

	v1.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper:   authenticate.SkipBasicAuth,
		Validator: authenticate.ValidateBasicAuth,
	}))

	v1.GET("/version", env.V1GetVersion)

	v1.POST("/register", env.V1PostRegister)

	// Route => handler
	v1.GET("/", env.V1GetRoot)
	v1.GET("/alist/:uuid", env.V1GetListByUUID)
	v1.GET("/alist/by/me", env.V1GetListsByMe)

	//e.POST("/alist/v1", env.V1PostAlist)
	//e.POST("/alist/v2", env.V1PostAlist)
	//e.POST("/alist/v3", env.V1PostAlist)
	//e.POST("/alist/v4", env.V1PostAlist)
	v1.POST("/alist", env.V1SaveAlist)
	v1.POST("/share/alist", env.V1ShareAlist)
	v1.PUT("/alist/:uuid", env.V1SaveAlist)
	v1.DELETE("/alist/:uuid", env.V1RemoveAlist)
	// Labels
	v1.POST("/labels", env.V1PostUserLabel)
	v1.GET("/labels/by/me", env.V1GetUserLabels)
	v1.DELETE("/labels/:uuid", env.V1RemoveUserLabel)

	// Start server
	listenOn := fmt.Sprintf(":%d", env.Port)
	e.Logger.Fatal(e.Start(listenOn))
}

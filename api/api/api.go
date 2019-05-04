package api

import (
	"fmt"
	"log"

	"github.com/freshteapot/learnalist-api/api/api/models"
	"github.com/freshteapot/learnalist-api/api/authenticate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Env exposing the data abstraction layer
type Env struct {
	Datastore    models.Datastore
	UserID       string
	Port         int
	DatabaseName string
	Dal          models.DAL
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
	db, err := models.NewDB(env.DatabaseName)
	if err != nil {
		log.Panic(err)
	}

	env.Datastore = &models.DAL{
		Db: db,
	}

	// Echo instance
	e := echo.New()
	e.HideBanner = true
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	authenticate.LookUp = env.Datastore.GetUserByCredentials
	e.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
		Skipper:   authenticate.SkipBasicAuth,
		Validator: authenticate.ValidateBasicAuth,
	}))

	e.GET("/version", env.GetVersion)

	e.POST("/register", env.PostRegister)

	// Route => handler
	e.GET("/", env.GetRoot)
	e.GET("/alist/:uuid", env.GetListByUUID)
	e.GET("/alist/by/me", env.GetListsByMe)

	//e.POST("/alist/v1", env.PostAlist)
	//e.POST("/alist/v2", env.PostAlist)
	//e.POST("/alist/v3", env.PostAlist)
	//e.POST("/alist/v4", env.PostAlist)
	e.POST("/alist", env.PostAlist)
	e.PUT("/alist/:uuid", env.PutAlist)
	e.DELETE("/alist/:uuid", env.RemoveAlist)

	// Start server
	listenOn := fmt.Sprintf(":%d", env.Port)
	e.Logger.Fatal(e.Start(listenOn))
}

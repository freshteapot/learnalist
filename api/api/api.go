package api

import (
	"fmt"
	"log"

	"github.com/freshteapot/learnalist/api/api/models"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	uuid "github.com/satori/go.uuid"
)

// Env exposing the data abstraction layer
type Env struct {
	Datastore    models.Datastore
	UserID       string
	Port         int
	DatabaseName string
	Dal          models.DAL
}

var basicAuth = ""

// @todo change this to be a configure file
var domain = "learnalist.net"

// UseBasicAuth Tell the api to use the following username:password.
func UseBasicAuth(auth string) {
	basicAuth = auth
}

// SetDomain set the domain this api is associated with.
func SetDomain(_domain string) {
	domain = _domain
}

// Return a new unique id :)
func getUUID() string {
	// @todo is this good enough?
	var secret = uuid.NewV4()
	u := uuid.NewV5(secret, domain)
	return u.String()
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
	// Gives pretty formatting
	// e.SetDebug(true)

	if basicAuth != "" {
		e.Use(middleware.BasicAuth(func(username, password string) bool {
			match := fmt.Sprintf("%s:%s", username, password)

			if match == basicAuth {
				return true
			}
			return false
		}))
	}

	// Middleware
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Route => handler
	e.GET("/", env.GetRoot)
	e.GET("/alist/:uuid", env.GetListByUUID)
	e.GET("/alist/by/:uuid", env.GetListsBy)

	e.POST("/alist", env.PostAlist)
	e.PUT("/alist/:uuid", env.PutAlist)
	e.DELETE("/alist/:uuid", env.RemoveAlist)

	// Start server
	listenOn := fmt.Sprintf(":%d", env.Port)
	e.Logger.Fatal(e.Start(listenOn))
}

package api

import (
	"fmt"
	"log"

	"github.com/freshteapot/learnalist/api/api/models"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	uuid "github.com/satori/go.uuid"
)

// Env exposing the data abstraction layer
type Env struct {
	db     models.Datastore
	userID string
}

type (
	responseMessage struct {
		Message string `json:"message"`
	}
)

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
func Run(port int, database string) {
	db, err := models.NewDB(database)
	if err != nil {
		log.Panic(err)
	}

	env := &Env{db, "me"}

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
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Route => handler
	e.GET("/", env.GetRoot)
	e.GET("/alist/:uuid", env.GetListByUUID)
	e.GET("/alist/by/:uuid", env.GetListsBy)

	e.POST("/alist", env.PostAlist)
	e.PUT("/alist/:uuid", env.PutAlist)
	e.PATCH("/alist/:uuid", env.PatchAlist)

	// Start server
	listenOn := fmt.Sprintf(":%d", port)
	e.Run(standard.New(listenOn))
}

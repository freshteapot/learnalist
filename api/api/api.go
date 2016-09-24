package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	uuid "github.com/satori/go.uuid"
)

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
	fmt.Println(domain)
	u := uuid.NewV5(secret, domain)
	return u.String()
}

// Run This starts the api listening on the port supplied
func Run(port int) {
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
	e.GET("/", func(c echo.Context) error {
		message := "1, 2, 3. Lets go!"
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusOK, response)
	})

	e.POST("/alist", func(c echo.Context) error {
		message := fmt.Sprintf("I want to upload alist with uuid: %s", getUUID())
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusOK, response)
	})

	e.PUT("/alist/:uuid", func(c echo.Context) error {
		uuid := c.Param("uuid")
		message := fmt.Sprintf("I want to alter alist with uuid: %s", uuid)
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusOK, response)
	})

	e.PATCH("/alist/:uuid", func(c echo.Context) error {
		uuid := c.Param("uuid")
		message := fmt.Sprintf("I want to alter alist with uuid: %s", uuid)
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusOK, response)
	})

	e.GET("/alist/:uuid", func(c echo.Context) error {
		uuid := c.Param("uuid")
		message := fmt.Sprintf("I want alist %s", uuid)
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusOK, response)
	})

	e.GET("/alist/by/:uuid", func(c echo.Context) error {
		uuid := c.Param("uuid")
		message := fmt.Sprintf("I want all lists by %s", uuid)
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusOK, response)
	})

	// Start server
	listenOn := fmt.Sprintf(":%d", port)
	e.Run(standard.New(listenOn))
}

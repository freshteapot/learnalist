package server

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	Port             int
	Domain           string
	CorsAllowOrigins string
	StaticSiteFolder string
	SiteCacheFolder  string
}

var server *echo.Echo
var config Config

func Init(_config Config) {
	config = _config
	server = echo.New()
	server.HideBanner = true
	// TODO change those which should be to Pre
	// Middleware
	server.Use(middleware.RequestID())
	server.Use(middleware.Logger())
	server.Use(middleware.Recover())

	server.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 9,
	}))

	// TODO can i use Pre(
	/*
		if config.CorsAllowOrigins != "" {
			allowOrigins := strings.Split(config.CorsAllowOrigins, ",")
			server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
				AllowOrigins: allowOrigins,
				AllowMethods: []string{http.MethodOptions, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
				AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderOrigin, echo.HeaderContentType},
			}))
		}
	*/
}

func Run() {
	// Start server
	listenOn := fmt.Sprintf(":%d", config.Port)
	server.Logger.Fatal(server.Start(listenOn))
}

package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	Port             string
	CorsAllowOrigins string
	HugoFolder       string
	SiteCacheFolder  string
}

var server *echo.Echo
var config Config

func Init(_config Config) {
	config = _config
	server = echo.New()
	server.HideBanner = true
	// Middleware
	server.Use(middleware.Recover())
	server.Use(middleware.Logger())
	server.Use(middleware.RequestID())

	server.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 9,
	}))

	if config.CorsAllowOrigins != "" {
		allowOrigins := strings.Split(config.CorsAllowOrigins, ",")
		server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: allowOrigins,
			AllowMethods: []string{http.MethodOptions, http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
			AllowHeaders: []string{echo.HeaderAuthorization, echo.HeaderOrigin, echo.HeaderContentType},
		}))
	}
}

func Run() {
	// Start server
	listenOn := fmt.Sprintf(":%s", config.Port)
	server.Logger.Fatal(server.Start(listenOn))
}

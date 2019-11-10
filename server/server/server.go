package server

import (
	"fmt"

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
}

func Run() {
	// Start server
	listenOn := fmt.Sprintf(":%s", config.Port)
	server.Logger.Fatal(server.Start(listenOn))
}

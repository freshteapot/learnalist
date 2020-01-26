package authenticate

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func Skip(c echo.Context) bool {
	url := c.Request().URL.Path
	method := c.Request().Method
	url = strings.TrimPrefix(url, "/api/v1")

	switch method {
	case http.MethodGet:
		if url == "/" {
			return true
		}

		if url == "/version" {
			return true
		}

		if strings.HasPrefix(url, "/oauth/") {
			return true
		}
	case http.MethodPost:
		// TODO Add a secret if you want to control who can register
		// Unfiltered ability to register a user
		if url == "/user/register" {
			return true
		}

		if url == "/user/login" {
			return true
		}
	default:
		return false
	}

	return false
}

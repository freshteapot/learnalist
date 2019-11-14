package authenticate

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func SkipAuth(c echo.Context) bool {
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
		// Unfiltered ability to register a user
		if url == "/register" {
			return true
		}
	default:
		return false
	}

	return false
}

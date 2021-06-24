package authenticate

import (
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/pkg/utils"
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

		if strings.Contains(url, "/plank/history/") {
			// A hack to try and get access to the plank/history/:uuid
			// If we add more than one, turn it into a filter
			authorization := c.Request().Header.Get("Authorization")
			if authorization == "" {
				//cookie.
				_, err := utils.GetCookieByName(c.Request().Cookies(), "x-authentication-bearer")
				if err != nil {
					return true
				}
			}
			return false
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

		if url == "/user/login/idp" {
			return true
		}
	default:
		return false
	}

	return false
}

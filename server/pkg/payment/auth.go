package payment

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/labstack/echo/v4"
)

func SkipAuth(c echo.Context) bool {
	url := c.Request().URL.Path
	method := c.Request().Method

	// We rely on the handler to verify the webhook
	if method == http.MethodPost && url == "/webhook" {
		return true
	}

	_, err := utils.GetCookieByName(c.Request().Cookies(), "x-authentication-bearer")
	if err != nil {
		return true
	}
	return false
}

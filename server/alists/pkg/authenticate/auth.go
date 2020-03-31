package authenticate

import (
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/labstack/echo/v4"
)

func SkipAuth(c echo.Context) bool {

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

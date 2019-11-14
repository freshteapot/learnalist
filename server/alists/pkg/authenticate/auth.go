package authenticate

import (
	"github.com/labstack/echo/v4"
)

func SkipAuth(c echo.Context) bool {
	authorization := c.Request().Header.Get("Authorization")
	if authorization == "" {
		return true
	}
	return false
}

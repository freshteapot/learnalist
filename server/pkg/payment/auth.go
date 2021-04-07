package payment

import (
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/labstack/echo/v4"
)

func SkipAuth(c echo.Context) bool {
	// TODO Skip if "success"
	// TODO Skip if "cancel"
	_, err := utils.GetCookieByName(c.Request().Cookies(), "x-authentication-bearer")
	if err != nil {
		return true
	}
	return false
}

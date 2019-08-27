package authenticate

import (
	"github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

var LookUp func(loginUser authenticate.LoginUser) (*uuid.User, error)

func SkipBasicAuth(c echo.Context) bool {
	authorization := c.Request().Header.Get("Authorization")
	if authorization == "" {
		return true
	}
	return false
}

func ValidateBasicAuth(username string, password string, c echo.Context) (bool, error) {
	loginUser := &authenticate.LoginUser{
		Username: username,
		Password: password,
	}
	user, err := LookUp(*loginUser)
	if err != nil {
		return true, nil
	}

	c.Set("loggedInUser", *user)
	return true, nil
}

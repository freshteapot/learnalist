package server

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/labstack/echo/v4"
)

func (m *Manager) Logout(c echo.Context) error {
	redirectURL := "/come-back-soon.html"
	loginCookie, err := c.Request().Cookie("x-authentication-bearer")
	if err != nil {
		if err == http.ErrNoCookie {
			return c.Redirect(http.StatusFound, redirectURL)
		}
	}

	token := loginCookie.Value
	session := m.Datastore.UserSession()
	authenticate.SendLogoutCookie(c)

	userUUID, err := session.GetUserUUIDByToken(token)
	if err != nil {
		fmt.Println("token not found, just redirect")
		return c.Redirect(http.StatusFound, redirectURL)
	}

	all := c.QueryParam("all")

	switch all {
	case "1":
		fmt.Println("Clear all sessions")
		session.RemoveSessionsForUser(userUUID)
	default:
		fmt.Println("Clear the sesion based on the token")
		session.RemoveSessionForUser(userUUID, token)
	}

	return c.Redirect(http.StatusFound, redirectURL)
}

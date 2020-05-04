package server

import (
	"fmt"
	"net/http"
	"time"

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
	cookie := authenticate.NewLoginCookie(token)
	cookie.Expires = time.Now().Add(-100 * time.Hour)
	cookie.MaxAge = -1
	cookie.Value = ""
	session := m.Datastore.UserSession()

	c.SetCookie(cookie)

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

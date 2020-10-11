package server

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
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

	eventKind := ""
	switch all {
	case "1":
		eventKind = event.KindUserLogoutSessions
		session.RemoveSessionsForUser(userUUID)
	default:
		eventKind = event.KindUserLogoutSession
		session.RemoveSessionForUser(userUUID, token)
	}

	event.GetBus().Publish(event.Eventlog{
		Kind: event.BrowserUserLogout,
		Data: event.EventUser{
			UUID: userUUID,
			Kind: eventKind,
		},
	})

	return c.Redirect(http.StatusFound, redirectURL)
}

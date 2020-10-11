package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func (m *Manager) V1OauthGoogleRedirect(c echo.Context) error {
	googleConfig := m.OauthHandlers.Google
	if googleConfig == nil {
		return c.String(http.StatusInternalServerError, "this website has not configured Google OAuth")
	}

	challenge, err := m.Datastore.UserSession().CreateWithChallenge()
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	url := googleConfig.AuthCodeURL(challenge, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusFound, url)
}

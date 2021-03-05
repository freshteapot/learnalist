package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func (s OauthService) V1OauthGoogleRedirect(c echo.Context) error {
	googleConfig := s.oauthHandlers.Google
	if googleConfig == nil {
		return c.String(http.StatusInternalServerError, "this website has not configured Google OAuth")
	}

	challenge, err := s.userSession.CreateWithChallenge()
	if err != nil {
		return c.String(http.StatusInternalServerError, i18n.InternalServerErrorFunny)
	}

	url := googleConfig.AuthCodeURL(challenge, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusFound, url)
}

package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
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
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	url := googleConfig.AuthCodeURL(challenge, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusFound, url)
}

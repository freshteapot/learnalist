package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func (s OauthService) V1OauthAppleIDRedirect(c echo.Context) error {
	oauthConfig := s.oauthHandlers.AppleID

	if oauthConfig == nil {
		return c.String(http.StatusInternalServerError, "this website has not configured AppleID OAuth")
	}

	challenge, err := s.userSession.CreateWithChallenge()
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	url := oauthConfig.AuthCodeURL(challenge, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusFound, url)
}

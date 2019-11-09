package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (m *Manager) V1OauthGoogleRedirect(c echo.Context) error {
	googleConfig := m.OauthHandlers.Google
	if googleConfig == nil {
		return c.String(http.StatusInternalServerError, "this website has not configured Google OAuth")
	}

	r := c.Request()
	token := r.FormValue("token")
	// Validate the token is in the process, by looking it up.
	url := googleConfig.AuthCodeURL(token)
	return c.Redirect(http.StatusFound, url)
}

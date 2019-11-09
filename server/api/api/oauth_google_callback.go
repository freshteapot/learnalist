package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func (m *Manager) V1OauthGoogleCallback(c echo.Context) error {
	googleConfig := m.OauthHandlers.Google
	r := c.Request()
	tempToken := r.FormValue("state")
	code := r.FormValue("code")
	// Add logic to handle looking up the state / token to see if it exists
	fmt.Println(tempToken)
	fmt.Println(code)
	token, err := googleConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Error: %s", err.Error()))
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		return c.String(http.StatusBadRequest, i18n.ErrorCannotReadResponse.Error())
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.String(http.StatusBadRequest, i18n.ErrorCannotReadResponse.Error())
	}
	fmt.Println(string(contents))
	user := make(map[string]interface{})
	if err := json.Unmarshal(contents, &user); err != nil {
		return c.String(http.StatusInternalServerError, i18n.ErrorInternal.Error())
	}

	if user["email"] == nil {
		return c.String(http.StatusBadRequest, "no email address returned by Google")
	}

	email := user["email"].(string)
	fmt.Println(email)

	return c.String(http.StatusOK, "Success")
}

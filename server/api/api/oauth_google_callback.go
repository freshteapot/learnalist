package api

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func (m *Manager) V1OauthGoogleCallback(c echo.Context) error {
	logger := logging.GetLogger()
	googleConfig := m.OauthHandlers.Google
	oauthHandler := m.Datastore.OAuthHandler()
	userSession := m.Datastore.UserSession()
	userFromIDP := m.Datastore.UserFromIDP()
	r := c.Request()
	challenge := r.FormValue("state")
	code := r.FormValue("code")
	// Confirm the challenge is valid
	has, err := userSession.IsChallengeValid(challenge)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Invalid challenge")
		response := HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	if !has {
		response := HttpResponseMessage{
			Message: "Invalid code / challenge, please try to login again",
		}
		return c.JSON(http.StatusBadRequest, response)
	}
	// Exchange the code for the token
	token, err := googleConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Error: %s", err.Error()))
	}
	// Lookup the user info
	ctx := context.Background()
	client := oauth2.NewClient(ctx, googleConfig.TokenSource(ctx, token))

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo?prettyprint=false", nil)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Talking to google failed")
		return c.String(http.StatusInternalServerError, i18n.ErrorInternal.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusBadRequest, i18n.ErrorCannotReadResponse.Error())
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.String(http.StatusBadRequest, i18n.ErrorCannotReadResponse.Error())
	}

	userInfo, err := oauth.GoogleConvertRawUserInfo(contents)
	if err != nil {
		// LOG the error
		return c.String(http.StatusBadRequest, "no email address returned by Google")
	}

	// Look up the user based on their email and association with google.
	userUUID, err := userFromIDP.Lookup("google", userInfo.Email)
	if err != nil {
		// Create a user
		userUUID, err = userFromIDP.Register("google", userInfo.Email, contents)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to register new user via idp")
			response := HttpResponseMessage{
				Message: i18n.InternalServerErrorFunny,
			}
			return c.JSON(http.StatusInternalServerError, response)
		}
	}
	// Create a session for the user
	userSessionToken := guuid.New()
	session := user.UserSession{
		Token:     userSessionToken.String(),
		UserUUID:  userUUID,
		Challenge: challenge,
	}

	err = userSession.Activate(session)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error("Failed to activate session")
		response := HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	// TODO save token info,
	// if RefreshToken == "" insert
	// if not
	if token.RefreshToken == "" {
		storedToken, err := oauthHandler.GetTokenInfo(string(userUUID))
		if err == nil {
			token.RefreshToken = storedToken.RefreshToken
		}
	}

	// TODO save token info
	oauthHandler.WriteTokenInfo(string(userUUID), token)

	/*
		b, _ := json.Marshal(token)
		fmt.Println(string(b))
		fmt.Println(token.AccessToken)
		rawIDToken := token.Extra("id_token").(string)
		fmt.Println(rawIDToken)
	*/
	// One way would be to post a cookie
	// Pass back a header
	// Have it in the payload, or in the actual html page for javascript to pick up and handle

	vars := make(map[string]interface{})
	vars["token"] = session.Token
	vars["refreshRedirectURL"] = "/welcome.html"
	vars["idp"] = "Google"

	var tpl bytes.Buffer
	oauthGoogleCallbackHtml200.Execute(&tpl, vars)

	return c.HTMLBlob(http.StatusOK, tpl.Bytes())
}

var oauthGoogleCallbackHtml200 = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head
	data-redirectUri="{{.refreshRedirectURL}}"
	data-token="{{.token}}"
>
<meta charset="utf-8" />
</head>
<body>
<h1>You have succesfully logged in via {{.idp}}</h1>
<p>You will now be redirected</p>
</body>
</html>
`))

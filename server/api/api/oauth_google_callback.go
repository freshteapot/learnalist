package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func (m *Manager) V1OauthGoogleCallback(c echo.Context) error {
	googleConfig := m.OauthHandlers.Google

	r := c.Request()
	challenge := r.FormValue("state")
	code := r.FormValue("code")

	has, err := m.Datastore.UserSession().IsChallengeValid(challenge)
	if err != nil {
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

	token, err := googleConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Sprintf("Error: %s", err.Error()))
	}

	ctx := context.Background()
	client := oauth2.NewClient(ctx, googleConfig.TokenSource(ctx, token))

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo?prettyprint=false", nil)
	if err != nil {
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

	// I will need a user before I can do a user_session or a oauth_xxx.
	// TODO Get user_uuid or make one,
	var userUUID user.UserUUID
	userUUID, err = m.Datastore.UserFromIDP().Lookup("google", "email", userInfo.Email)
	if err != nil {
		// Create a user
		userUUID, err = m.Datastore.UserFromIDP().Register("google", "email", userInfo.Email, contents)
	}

	userSessionToken := guuid.New()
	userSession := user.UserSession{
		Token:     userSessionToken.String(),
		UserUUID:  userUUID,
		Challenge: challenge,
	}

	err = m.Datastore.UserSession().Activate(userSession)
	if err != nil {
		response := HttpResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	// TODO save token info,
	// if RefreshToken == "" insert
	// if not
	if token.RefreshToken != "" {
		// TODO save token info
	}

	/*
		b, _ := json.Marshal(token)
		fmt.Println(string(b))
		fmt.Println(token.AccessToken)
		rawIDToken := token.Extra("id_token").(string)
		fmt.Println(rawIDToken)
	*/

	return c.String(http.StatusOK, "Todo how to share the session token")
}

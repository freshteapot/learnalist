package api

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	guuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func (m *Manager) V1OauthGoogleCallback(c echo.Context) error {
	logger := m.logger
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
		return c.String(http.StatusInternalServerError, i18n.InternalServerErrorFunny)
	}

	if !has {
		return c.String(http.StatusBadRequest, "Invalid code / challenge, please try to login again")
	}

	// Exchange the code for the token
	token, err := googleConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		// Log
		response := "Exhange of code to token failed"
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Error(response)

		return c.String(http.StatusBadRequest, response)
	}

	ctx := context.Background()
	client := googleConfig.Client(ctx, token)

	// Have read the code, its very unlikely this would throw err (famous last words)
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo?prettyprint=false", nil)
	resp, err := client.Do(req)
	if err != nil {
		return c.String(http.StatusBadRequest, "Something went wrong, please try again")
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.String(http.StatusBadRequest, "Something went wrong, please try again")
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "Something went wrong, please try again")
	}

	userInfo, err := oauth.GoogleConvertRawUserInfo(contents)
	if err != nil {
		// LOG the error
		return c.String(http.StatusBadRequest, "no email address returned by Google")
	}

	// Look up the user based on their email and association with google.
	userUUID, err := userFromIDP.Lookup(user.IDPKeyGoogle, user.IDPKindUserID, userInfo.ID)
	if err != nil {
		if err != utils.ErrNotFound {
			logger.WithFields(logrus.Fields{
				"event": "idp-lookup-user-info",
				"error": err,
			}).Error("Issue in google callback")
			return c.String(http.StatusBadRequest, "Something went wrong, please try again")
		}

		// Create a user
		userUUID, err = userFromIDP.Register(user.IDPKeyGoogle, user.IDPKindUserID, userInfo.ID, contents)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"event": "idp-register-user",
				"error": err,
			}).Error("Failed to register new user via idp")
			return c.String(http.StatusInternalServerError, i18n.InternalServerErrorFunny)
		}

		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: event.ApiUserRegister,
			Data: event.EventUser{
				UUID: userUUID,
				Kind: event.KindUserRegisterIDPGoogle,
			},
		})

		// Write an empty list
		m.HugoHelper.WriteListsByUser(userUUID, m.Datastore.GetAllListsByUser(userUUID))
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
			"event": "idp-session-activate",
			"error": err,
		}).Error("Failed to activate session")
		return c.String(http.StatusInternalServerError, i18n.InternalServerErrorFunny)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiUserLogin,
		Data: event.EventUser{
			UUID: userUUID,
			Kind: event.KindUserLoginIDPGoogle,
		},
	})

	// If refreshToken is empty, we look it up in the db
	// before we write it back to the db.
	if token.RefreshToken == "" {
		storedToken, err := oauthHandler.GetTokenInfo(string(userUUID))
		if err == nil {
			token.RefreshToken = storedToken.RefreshToken
		}
	}

	err = oauthHandler.WriteTokenInfo(string(userUUID), token)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"event": "idp-session-activate",
			"error": err,
		}).Error("Failed to save token info")
		return c.String(http.StatusInternalServerError, i18n.InternalServerErrorFunny)
	}

	vars := make(map[string]interface{})
	vars["token"] = session.Token
	vars["userUUID"] = userUUID
	vars["refreshRedirectURL"] = "/welcome.html"
	vars["idp"] = user.IDPKeyGoogle

	var tpl bytes.Buffer
	oauthCallbackHtml200.Execute(&tpl, vars)

	cookie := authenticate.NewLoginCookie(session.Token)
	c.SetCookie(cookie)
	return c.HTMLBlob(http.StatusOK, tpl.Bytes())
}

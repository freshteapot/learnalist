package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// @event.emit: event.ApiUserRegister
// @event.emit: event.ApiUserLogin
func (s UserService) LoginViaIDP(c echo.Context) error {
	userFromIDP := s.userFromIDP
	userSession := s.userSession

	var input openapi.HttpUserLoginIdpInput
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.HTTPResponseMessage{
			Message: "Check the documentation",
		})
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"idp": input.Idp,
	})

	allowedIdps := s.oauthHandlers.Keys()

	if !utils.StringArrayContains(allowedIdps, input.Idp) {
		logContext.WithFields(logrus.Fields{
			"event": "idp-not-supported",
			"error": err,
		}).Error("Future feature")

		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: fmt.Sprintf("Idp not supported: %s", strings.Join(allowedIdps, ",")),
		})
	}

	// Convert token
	//var token *Tokeninfo
	var (
		extUserUUID                                 string
		eventKindLoginValidToken, eventKindRegister string
	)
	// This can be refactored, by looking at the aud
	// Or keeping the idp and having the function
	switch input.Idp {
	case oauth.IDPKeyGoogle:
		extUserUUID, err = s.oauthHandlers.Google.GetUserUUIDFromIDP(input)
		eventKindRegister = event.KindUserRegisterIDPGoogle
		eventKindLoginValidToken = event.KindUserLoginIDPGoogleViaIdToken
	case oauth.IDPKeyApple:
		extUserUUID, err = s.oauthHandlers.AppleID.GetUserUUIDFromIDP(input)
		eventKindRegister = event.KindUserRegisterIDPApple
		eventKindLoginValidToken = event.KindUserLoginIDPAppleViaIdToken
	default:
		logContext.WithFields(logrus.Fields{
			"event": "idp-not-supported-2",
			"error": err,
		}).Error("Future feature")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event":  "idp-token-verification",
			"method": "GetUserUUIDFromIDP",
			"error":  err,
		}).Error("Issue in login via idp")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	userUUID, err := userFromIDP.Lookup(input.Idp, user.IDPKindUserID, extUserUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"event": "idp-lookup-user-info",
				"error": err,
			}).Error("Issue in login via idp")
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}

		// Create a new user
		userUUID, err = userFromIDP.Register(input.Idp, user.IDPKindUserID, extUserUUID, []byte(``))
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"event": "idp-register-user",
				"error": err,
			}).Error("Failed to register new user via login with id_token")
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}

		pref := user.UserPreference{}
		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: event.ApiUserRegister,
			Data: event.EventNewUser{
				UUID: userUUID,
				Kind: eventKindRegister,
				Data: pref,
			},
		})
	}

	session, err := userSession.NewSession(userUUID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event": "idp-session-create",
			"idp":   input.Idp,
			"error": err,
		}).Error("Failed to create session")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiUserLogin,
		Data: event.EventUser{
			UUID: userUUID,
			Kind: eventKindLoginValidToken,
		},
	})

	return c.JSON(http.StatusOK, openapi.HttpUserLoginResponse{
		Token:    session.Token,
		UserUuid: userUUID,
	})
}

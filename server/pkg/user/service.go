package user

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type UserService struct {
	db            *sqlx.DB
	userFromIDP   UserFromIDP
	userSession   Session
	hugoHelper    hugo.HugoHelper
	oauthHandlers oauth.Handlers
	logContext    logrus.FieldLogger
}

// @openapi.path.tag: user
func NewService(db *sqlx.DB,
	oauthHandlers oauth.Handlers,
	userFromIDP UserFromIDP,
	userSession Session,
	hugoHelper hugo.HugoHelper,
	logContext logrus.FieldLogger,
) UserService {
	return UserService{
		db:            db,
		oauthHandlers: oauthHandlers,
		userFromIDP:   userFromIDP,
		userSession:   userSession,
		hugoHelper:    hugoHelper,
		logContext:    logContext,
	}
}

// @event.emit: event.ApiUserRegister
// @event.emit: event.ApiUserLogin
func (s UserService) LoginViaIDP(c echo.Context) error {
	userFromIDP := s.userFromIDP
	userSession := s.userSession

	logContext := s.logContext
	var input oauth.IDPOauthInput
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	idpAllowed := s.oauthHandlers.Keys()

	if !utils.StringArrayContains(idpAllowed, input.Idp) {
		logContext.WithFields(logrus.Fields{
			"event": "idp-not-supported",
			"idp":   input.Idp,
			"error": err,
		}).Error("Future feature")
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
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
			"idp":   input.Idp,
			"error": err,
		}).Error("Future feature")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event":  "idp-token-verification",
			"method": "GetUserUUIDFromIDP",
			"idp":    input.Idp,
			"error":  err,
		}).Error("Issue in login via idp")
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	userUUID, err := userFromIDP.Lookup(input.Idp, IDPKindUserID, extUserUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"event": "idp-lookup-user-info",
				"idp":   input.Idp,
				"error": err,
			}).Error("Issue in login via idp")
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}

		// Create a new user
		userUUID, err = userFromIDP.Register(input.Idp, IDPKindUserID, extUserUUID, []byte(``))
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"event": "idp-register-user",
				"idp":   input.Idp,
				"error": err,
			}).Error("Failed to register new user via login with id_token")
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}

		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: event.ApiUserRegister,
			Data: event.EventUser{
				UUID: userUUID,
				Kind: eventKindRegister,
			},
		})

		// TODO use event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{})
		// Write an empty list
		lists := make([]alist.ShortInfo, 0)
		s.hugoHelper.WriteListsByUser(userUUID, lists)
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

	// TODO is this kind supported in slack? (confirm)
	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiUserLogin,
		Data: event.EventUser{
			UUID: userUUID,
			Kind: eventKindLoginValidToken,
		},
	})

	return c.JSON(http.StatusOK, api.HTTPLoginResponse{
		Token:    session.Token,
		UserUUID: userUUID,
	})
}

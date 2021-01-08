package user

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	guuid "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/oauth2/v1"
)

/*
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/login/idp' -d'
{
    "idp": "google",
	"id_token": "XXX",
	"access_token": "XXX",
}
'
*/
type UserService struct {
	db          *sqlx.DB
	userFromIDP UserFromIDP
	userSession Session
	hugoHelper  hugo.HugoHelper
	issuedTo    []string
	logContext  logrus.FieldLogger
}

type HTTPUserLoginIDPInput struct {
	Idp         string `json:"idp"`
	IDToken     string `json:"id_token"`
	AccessToken string `json:"access_token"`
}

func NewService(db *sqlx.DB, issuedTo []string, userFromIDP UserFromIDP, userSession Session, hugoHelper hugo.HugoHelper, logContext logrus.FieldLogger) UserService {
	return UserService{
		db:          db,
		userFromIDP: userFromIDP,
		userSession: userSession,
		hugoHelper:  hugoHelper,
		issuedTo:    issuedTo,
		logContext:  logContext,
	}
}

func (s UserService) LoginViaIDP(c echo.Context) error {
	userFromIDP := s.userFromIDP
	userSession := s.userSession

	logContext := s.logContext
	var input HTTPUserLoginIDPInput
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	if input.Idp != "google" {
		logContext.WithFields(logrus.Fields{
			"event": "idp-not-supported",
			"idp":   input.Idp,
			"error": err,
		}).Error("Future feature")
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	// Convert token
	token, err := verifyIdToken(input.IDToken)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event": "idp-token-verification",
			"idp":   input.Idp,
			"error": err,
		}).Error("Issue in login via idp")
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	if !utils.StringArrayContains(s.issuedTo, token.IssuedTo) {
		logContext.WithFields(logrus.Fields{
			"event":           "idp-issued",
			"idp":             input.Idp,
			"input_issued_to": token.IssuedTo,
			"error":           err,
		}).Error("Issue in login via idp")
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	extUserID := token.UserId
	userUUID, err := userFromIDP.Lookup(input.Idp, IDPKindUserID, extUserID)
	if err != nil {
		if err != utils.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"event": "idp-lookup-user-info",
				"idp":   input.Idp,
				"error": err,
			}).Error("Issue in login via idp")
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}

		// TODO needs better http setup
		req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo?prettyprint=false", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", input.AccessToken))
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"event": "idp-user-info-via-google-1",
				"error": err,
			}).Error("Issue in login via idp")
			return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
		}

		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logContext.WithFields(logrus.Fields{
				"event":       "idp-user-info-via-google-2",
				"status_code": resp.StatusCode,
				"error":       err,
			}).Error("Issue in login via idp")
			return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
		}

		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"event": "idp-user-info-via-google-3",
				"error": err,
			}).Error("Issue in login via idp")
			return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
		}

		// Create a new user
		userUUID, err = userFromIDP.Register(input.Idp, IDPKindUserID, extUserID, contents)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"event": "idp-register-user",
				"idp":   input.Idp,
				"error": err,
			}).Error("Failed to register new user via login with id_token")
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}
		/*
			logContext.WithFields(logrus.Fields{
				"event":              "idp-lookup-user-info",
				"idp":                input.Idp,
				"verified_email":     token.VerifiedEmail,
				"internal_user_uuid": userUUID,
				"external_user_id":   token.UserId,
				"email":              token.Email,
			}).Info("user")
		*/
		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: event.ApiUserRegister,
			Data: event.EventUser{
				UUID: userUUID,
				Kind: event.KindUserRegisterIDPGoogle,
			},
		})

		// TODO move to event
		// Write an empty list
		lists := make([]alist.ShortInfo, 0)
		s.hugoHelper.WriteListsByUser(userUUID, lists)
	}

	// Create a session for the user
	userSessionToken := guuid.New()
	// Hack to make it work
	challenge, _ := userSession.CreateWithChallenge()
	session := UserSession{
		Token:     userSessionToken.String(),
		UserUUID:  userUUID,
		Challenge: challenge,
	}

	err = userSession.Activate(session)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event": "idp-session-activate",
			"idp":   input.Idp,
			"error": err,
		}).Error("Failed to activate session")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	// TODO is this kind supported in slack?
	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiUserLogin,
		Data: event.EventUser{
			UUID: userUUID,
			Kind: event.KindUserLoginIDPGoogleViaIdToken,
		},
	})

	return c.JSON(http.StatusOK, api.HTTPLoginResponse{
		Token:    session.Token,
		UserUUID: userUUID,
	})
}

var httpClient = &http.Client{}

func verifyIdToken(idToken string) (*oauth2.Tokeninfo, error) {
	oauth2Service, err := oauth2.New(httpClient)
	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.IdToken(idToken)

	tokenInfo, err := tokenInfoCall.Do()
	if err != nil {
		return nil, err
	}
	return tokenInfo, nil
}

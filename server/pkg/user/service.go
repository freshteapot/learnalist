package user

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	guuid "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/oauth2/v1"
)

type UserService struct {
	db          *sqlx.DB
	userFromIDP UserFromIDP
	userSession Session
	hugoHelper  hugo.HugoHelper
	issuedTo    []string
	logContext  logrus.FieldLogger
}

type LoginIDPInput struct {
	Idp     string `json:"idp"`
	IDToken string `json:"id_token"`
	Via     string `json:"via"`
}

type LoginIDPResponse struct {
	Token    string `json:"token"`
	UserUUID string `json:"user_uuid"`
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
	var input LoginIDPInput
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {

		return c.JSON(http.StatusBadRequest, api.HTTPResponseMessage{
			Message: "TODO 1",
		})
	}

	if input.Idp != "google" {
		return c.JSON(http.StatusBadRequest, api.HTTPResponseMessage{
			Message: "TODO 2",
		})
	}

	if input.Via != "plank.app.v1" {
		return c.JSON(http.StatusBadRequest, api.HTTPResponseMessage{
			Message: "TODO 2",
		})
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

	// TODO pass this in
	if !utils.StringArrayContains(s.issuedTo, token.IssuedTo) {
		logContext.WithFields(logrus.Fields{
			"event":           "idp-issued",
			"idp":             input.Idp,
			"input_issued_to": token.IssuedTo,
			"error":           err,
		}).Error("Issue in login via idp")
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	//fmt.Println("VerifiedEmail", token.VerifiedEmail)
	//fmt.Println("UserId", token.UserId)
	//fmt.Println("Email", token.Email)

	// Lookup user by idp + ID
	contents := []byte(`{"action":"//TODO"}`)
	extUserID := token.UserId
	userUUID, err := userFromIDP.Lookup(input.Idp, IDPKindUserID, extUserID)
	if err != nil {
		if err != ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"event": "idp-lookup-user-info",
				"idp":   input.Idp,
				"error": err,
			}).Error("Issue in login via idp")
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}
		// TODO what should contents be?
		// TODO add IDPKindUserID to register
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

		event.GetBus().Publish(event.Eventlog{
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
	event.GetBus().Publish(event.Eventlog{
		Kind: event.ApiUserLogin,
		Data: event.EventUser{
			UUID: userUUID,
			Kind: event.KindUserLoginIDPGoogleViaIdToken,
		},
	})

	return c.JSON(http.StatusOK, LoginIDPResponse{
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

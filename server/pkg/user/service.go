package user

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	guuid "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/tideland/gorest/jwt"
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
	AccessToken string `json:"access_token"` // TODO remove
	Code        string `json:"code"`
}

// @openapi.path.tag: user
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

// @event.emit: event.ApiUserRegister
// @event.emit: event.ApiUserLogin
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

	// TODO change to include apple
	// TODO make configurable
	idpAllowed := []string{
		IDPKeyGoogle,
		IDPKeyApple,
	}

	if !utils.StringArrayContains(idpAllowed, input.Idp) {
		logContext.WithFields(logrus.Fields{
			"event": "idp-not-supported",
			"idp":   input.Idp,
			"error": err,
		}).Error("Future feature")
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	// Convert token
	var token *Tokeninfo

	switch input.Idp {
	case IDPKeyGoogle:
		token, err = verifyIdTokenGoogle(input.IDToken)
	case IDPKeyApple:
		// TODO need the code or I just validate?
		token, err = verifyIDTokenApple(input.IDToken)
	}

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event": "idp-token-verification",
			"idp":   input.Idp,
			"error": err,
		}).Error("Issue in login via idp")
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	if !utils.StringArrayContains(s.issuedTo, token.Aud) {
		logContext.WithFields(logrus.Fields{
			"event":           "idp-issued",
			"idp":             input.Idp,
			"input_issued_to": token.Aud,
			"error":           err,
		}).Error("Issue in login via idp")
		return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
	}

	extUserID := token.Sub
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

		// TODO idp specific
		// TODO needs better http setup
		//req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo?prettyprint=false", nil)
		//req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", input.AccessToken))
		//resp, err := http.DefaultClient.Do(req)
		//if err != nil {
		//	logContext.WithFields(logrus.Fields{
		//		"event": "idp-user-info-via-google-1",
		//		"error": err,
		//	}).Error("Issue in login via idp")
		//	return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
		//}
		//
		//defer resp.Body.Close()
		//if resp.StatusCode != http.StatusOK {
		//	logContext.WithFields(logrus.Fields{
		//		"event":       "idp-user-info-via-google-2",
		//		"status_code": resp.StatusCode,
		//		"error":       err,
		//	}).Error("Issue in login via idp")
		//	return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
		//}
		//
		//contents, err := ioutil.ReadAll(resp.Body)
		//if err != nil {
		//	logContext.WithFields(logrus.Fields{
		//		"event": "idp-user-info-via-google-3",
		//		"error": err,
		//	}).Error("Issue in login via idp")
		//	return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
		//}
		contents := []byte(``)

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

type Tokeninfo struct {
	// Audience: Who is the intended audience for this token. In general the
	// same as issued_to.
	Aud string `json:"aud,omitempty"`
	// UserId: The obfuscated user id.
	Sub string `json:"sub,omitempty"`
}

//
// TODO pass in idp?
func verifyIdTokenGoogle(idToken string) (*Tokeninfo, error) {
	// TODO this is google specific
	oauth2Service, err := oauth2.New(httpClient)
	tokenInfoCall := oauth2Service.Tokeninfo()
	tokenInfoCall.IdToken(idToken)

	tokenInfo, err := tokenInfoCall.Do()
	if err != nil {
		return nil, err
	}

	return &Tokeninfo{
		Aud: tokenInfo.Audience,
		Sub: tokenInfo.UserId,
	}, nil
}

func verifyIDTokenApple(idToken string) (*Tokeninfo, error) {
	// Or we go with "github.com/Timothylock/go-signin-with-apple/apple"
	j, err := jwt.Decode(idToken)
	if err != nil {
		return nil, errors.New("bad token")
	}

	leeway := time.Minute
	if !j.IsValid(leeway) {
		return nil, errors.New("time has passed")
	}

	iss, _ := j.Claims().GetString("iss")

	if iss != "https://appleid.apple.com" {
		return nil, errors.New("bad-issuer")
	}

	aud, _ := j.Claims().GetString("aud")
	sub, _ := j.Claims().GetString("sub")

	return &Tokeninfo{
		Aud: aud,
		Sub: sub,
	}, nil
}

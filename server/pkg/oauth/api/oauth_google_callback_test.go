package api_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	oauthApi "github.com/freshteapot/learnalist-api/server/pkg/oauth/api"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

var _ = Describe("Testing Google Oauth callback", func() {

	var (
		logger *logrus.Logger
		hook   *test.Hook

		method    string
		uriPrefix string

		c   echo.Context
		e   *echo.Echo
		req *http.Request
		rec *httptest.ResponseRecorder

		hugoHelper   *mocks.HugoSiteBuilder
		userSession  *mocks.Session
		userFromIDP  *mocks.UserFromIDP
		oauth2Config *mocks.OAuth2ConfigInterface
		service      oauthApi.OauthService

		challenge     string
		fakeExtUserID string
		want          error
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()

		userFromIDP = &mocks.UserFromIDP{}
		userManagement := &mocks.Management{}
		userSession = &mocks.Session{}
		oauth2Config = &mocks.OAuth2ConfigInterface{}
		oauthHandlers := oauth.Handlers{
			Google: oauth2Config,
		}

		hugoHelper = &mocks.HugoSiteBuilder{}

		service = oauthApi.NewService(
			userManagement,
			hugoHelper,
			oauthHandlers,
			userSession,
			userFromIDP,
			logger)

		method = http.MethodGet
		uriPrefix = "/api/v1/oauth/google/callback"
		e = echo.New()

		challenge = "fake-123"
		fakeExtUserID = "fake-ext-user-id-123"
		want = errors.New("fail")
	})

	When("On return from the idp, we check the challenge is valid", func() {
		It("An error whilst looking up the challenge", func() {
			code := ""
			challenge := "fake-123"
			uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, code)
			userSession.On("IsChallengeValid", challenge).Return(false, want)

			req, rec = testutils.SetupJSONEndpoint(method, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath(uri)

			service.V1OauthGoogleCallback(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`Sadly, our service has taken a nap.`))
		})

		It("Challenge is not valid", func() {
			code := ""
			challenge := "fake-123"
			uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, code)

			userSession.On("IsChallengeValid", challenge).Return(false, nil)
			req, rec = testutils.SetupJSONEndpoint(method, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath(uri)

			service.V1OauthGoogleCallback(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`Invalid code / challenge, please try to login again`))
		})
	})

	It("Handle when the exchange fails", func() {
		userSession.On("IsChallengeValid", challenge).Return(true, nil)
		oauth2Config.On("Exchange", mock.Anything, mock.Anything).Return(nil, want)

		uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
		req, rec = testutils.SetupJSONEndpoint(method, uri, "")
		c = e.NewContext(req, rec)
		c.SetPath(uri)

		service.V1OauthGoogleCallback(c)
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
		Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`Exhange of code to token failed`))
		Expect(hook.LastEntry().Data["error"]).To(Equal(want))
		Expect(hook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
	})

	It("Talking to the idp, we fail to get the userUUID", func() {
		input := openapi.HttpUserLoginIdpInput{
			Idp:     oauth.IDPKeyGoogle,
			IdToken: "fake-token",
		}

		token := &oauth2.Token{}
		token = token.WithExtra(map[string]interface{}{
			"id_token": "fake-token",
			"aud":      "fake-aud",
			"sub":      "fake-user-123",
			"iss":      "fake-iss",
		})

		userSession.On("IsChallengeValid", challenge).Return(true, nil)
		userFromIDP.On("Lookup", oauth.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID).Return("", want)
		oauth2Config.On("Exchange", mock.Anything, mock.Anything).Return(token, nil)
		oauth2Config.On("GetUserUUIDFromIDP", input).Return("", want)
		uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
		req, rec = testutils.SetupJSONEndpoint(method, uri, "")
		c = e.NewContext(req, rec)
		c.SetPath(uri)

		service.V1OauthGoogleCallback(c)
		Expect(rec.Code).To(Equal(http.StatusForbidden))
		testutils.CheckMessageResponseFromResponseRecorder(rec, api.HTTPAccessDeniedResponse.Message)

	})

	When("Looking up the user in the system", func() {
		BeforeEach(func() {
			uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
			req, rec = testutils.SetupJSONEndpoint(method, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath(uri)

			userSession.On("IsChallengeValid", challenge).Return(true, nil)
			input := openapi.HttpUserLoginIdpInput{
				Idp:     oauth.IDPKeyGoogle,
				IdToken: "fake-token",
			}

			token := &oauth2.Token{}
			token = token.WithExtra(map[string]interface{}{
				"id_token": "fake-token",
				"aud":      "fake-aud",
				"sub":      fakeExtUserID,
				"iss":      "fake-iss",
			})

			oauth2Config.On("Exchange", mock.Anything, mock.Anything).Return(token, nil)
			oauth2Config.On("GetUserUUIDFromIDP", input).Return(fakeExtUserID, nil)
		})

		It("User found, but fail to create a session due to storage", func() {
			userFromIDP.On("Lookup", oauth.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID).Return("fake-user-123", nil)
			userSession.On("Activate", mock.Anything).Return(want)
			service.V1OauthGoogleCallback(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			Expect(hook.LastEntry().Data["event"]).To(Equal("idp-session-activate"))
		})

		When("User lookup returns not found, we register the user", func() {
			It("Failed to lookup user due to storage", func() {
				userFromIDP.On("Lookup", oauth.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID).Return("", want)
				service.V1OauthGoogleCallback(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(hook.LastEntry().Data["event"]).To(Equal("idp-lookup-user-info"))
			})

			It("Failed to register user due to saving to storage", func() {
				userFromIDP.On("Lookup", oauth.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID).Return("", utils.ErrNotFound)
				userFromIDP.On("Register", oauth.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID, mock.Anything).Return("", errors.New("fail"))
				uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
				req, rec = testutils.SetupJSONEndpoint(method, uri, "")
				c = e.NewContext(req, rec)
				c.SetPath(uri)

				service.V1OauthGoogleCallback(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(hook.LastEntry().Data["event"]).To(Equal("idp-register-user"))
			})

			It("Success, user registered and now the post register step(s)", func() {
				userUUID := "fake-uuid-123"
				noLists := make([]alist.ShortInfo, 0)

				hugoHelper.On("WriteListsByUser", userUUID, noLists)
				userFromIDP.On("Lookup", oauth.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID).Return("", utils.ErrNotFound)
				userFromIDP.On("Register", oauth.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID, mock.Anything).Return(userUUID, nil)

				userSession.On("Activate", mock.Anything).Return(nil)

				// I bet there is a better way
				try := 0
				eventMessageBus := &mocks.EventlogPubSub{}
				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					if try == 0 {
						Expect(moment.Kind).To(Equal(event.ApiUserRegister))
						Expect(moment.Data.(event.EventUser).Kind).To(Equal(event.KindUserRegisterIDPGoogle))
						try = 1
						return true
					}

					if try == 1 {
						Expect(moment.Kind).To(Equal(event.ApiUserLogin))
						Expect(moment.Data.(event.EventUser).Kind).To(Equal(event.KindUserLoginIDPGoogle))
						return true
					}

					return true
				}))

				event.SetBus(eventMessageBus)

				service.V1OauthGoogleCallback(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				// Check the cookie exists
				_, err := utils.GetCookieByName(rec.Result().Cookies(), "x-authentication-bearer")
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})

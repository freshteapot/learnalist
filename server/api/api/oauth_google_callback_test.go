package api_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/api"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
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
		logger          *logrus.Logger
		hook            *test.Hook
		datastore       *mocks.Datastore
		method          string
		uriPrefix       string
		e               *echo.Echo
		userSession     *mocks.Session
		userFromIDP     *mocks.UserFromIDP
		oauthReadWriter *mocks.OAuthReadWriter
		oauth2Config    *mocks.OAuth2ConfigInterface
		manager         *api.Manager
		challenge       string
		fakeExtUserID   string
		want            error
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		datastore = &mocks.Datastore{}
		userFromIDP = &mocks.UserFromIDP{}
		userManagement := &mocks.Management{}
		userSession = &mocks.Session{}
		oauthReadWriter = &mocks.OAuthReadWriter{}
		oauth2Config = &mocks.OAuth2ConfigInterface{}
		acl := &mocks.Acl{}
		oauthHandlers := oauth.Handlers{
			Google: oauth2Config,
		}

		testHugoHelper := &mocks.HugoSiteBuilder{}
		testHugoHelper.On("WriteListsByUser", mock.Anything, mock.Anything)
		datastore.On("UserSession").Return(userSession)
		datastore.On("UserFromIDP").Return(userFromIDP)
		datastore.On("OAuthHandler").Return(oauthReadWriter)

		manager = api.NewManager(datastore, userManagement, acl, "", testHugoHelper, oauthHandlers, "", logger)

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
			req, rec := setupFakeEndpoint(method, uri, "")
			c := e.NewContext(req, rec)

			manager.V1OauthGoogleCallback(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`Sadly, our service has taken a nap.`))
		})

		It("Challenge is not valid", func() {
			code := ""
			challenge := "fake-123"
			uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, code)

			userSession.On("IsChallengeValid", challenge).Return(false, nil)
			req, rec := setupFakeEndpoint(method, uri, "")
			c := e.NewContext(req, rec)

			manager.V1OauthGoogleCallback(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`Invalid code / challenge, please try to login again`))
		})
	})

	It("Handle when the exchange fails", func() {
		userSession.On("IsChallengeValid", challenge).Return(true, nil)
		oauth2Config.On("Exchange", mock.Anything, mock.Anything).Return(nil, want)

		uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
		req, rec := setupFakeEndpoint(method, uri, "")
		c := e.NewContext(req, rec)
		manager.V1OauthGoogleCallback(c)
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
		Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`Exhange of code to token failed`))
		Expect(hook.LastEntry().Data["error"]).To(Equal(want))
		Expect(hook.LastEntry().Level).To(Equal(logrus.ErrorLevel))
	})

	When("Looking up the user info from google", func() {
		It("Google returns non 200", func() {
			client := testutils.NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 400,
					Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
					Header:     make(http.Header),
				}
			})
			userSession.On("IsChallengeValid", challenge).Return(true, nil)
			oauth2Config.On("Exchange", mock.Anything, mock.Anything).Return(&oauth2.Token{}, nil)
			oauth2Config.On("Client", mock.Anything, mock.Anything).Return(client)

			uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
			req, rec := setupFakeEndpoint(method, uri, "")
			c := e.NewContext(req, rec)
			manager.V1OauthGoogleCallback(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
		})

		It("Response is not valid json", func() {
			client := testutils.NewTestClient(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewBufferString("")),
					Header:     make(http.Header),
				}
			})
			userSession.On("IsChallengeValid", challenge).Return(true, nil)
			oauth2Config.On("Exchange", mock.Anything, mock.Anything).Return(&oauth2.Token{}, nil)
			oauth2Config.On("Client", mock.Anything, mock.Anything).Return(client)

			uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
			req, rec := setupFakeEndpoint(method, uri, "")
			c := e.NewContext(req, rec)
			manager.V1OauthGoogleCallback(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal("no email address returned by Google"))
		})
	})

	When("Looking up the user in the system", func() {
		BeforeEach(func() {
			client := testutils.NewTestClient(func(req *http.Request) *http.Response {
				body := testutils.GetTestDataAsJSONOneline("idp-google-user-info.json")
				return &http.Response{
					StatusCode: 200,
					Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
					Header:     make(http.Header),
				}
			})
			userSession.On("IsChallengeValid", challenge).Return(true, nil)
			oauth2Config.On("Exchange", mock.Anything, mock.Anything).Return(&oauth2.Token{}, nil)
			oauth2Config.On("Client", mock.Anything, mock.Anything).Return(client)
		})

		It("Response is not valid json", func() {
			want := errors.New("fail")
			userFromIDP.On("Lookup", user.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID).Return("", want)
			uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
			req, rec := setupFakeEndpoint(method, uri, "")
			c := e.NewContext(req, rec)
			manager.V1OauthGoogleCallback(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
		})

		When("User lookup returns not found, we register the user", func() {
			It("Failed to register user due to saving to storage", func() {
				userFromIDP.On("Lookup", user.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID).Return("", utils.ErrNotFound)
				userFromIDP.On("Register", user.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID, mock.Anything).Return("", errors.New("fail"))
				uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
				req, rec := setupFakeEndpoint(method, uri, "")
				c := e.NewContext(req, rec)
				manager.V1OauthGoogleCallback(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(hook.LastEntry().Data["event"]).To(Equal("idp-register-user"))
			})

			It("Success, user registered and now the post register step(s)", func() {
				userUUID := "fake-uuid-123"
				noLists := make([]alist.ShortInfo, 0)

				testHugoHelper := &mocks.HugoSiteBuilder{}
				testHugoHelper.On("WriteListsByUser", mock.Anything, mock.Anything)
				userFromIDP.On("Lookup", user.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID).Return("", utils.ErrNotFound)
				userFromIDP.On("Register", user.IDPKeyGoogle, user.IDPKindUserID, fakeExtUserID, mock.Anything).Return(userUUID, nil)
				datastore.On("GetAllListsByUser", userUUID).Return(noLists)
				userSession.On("Activate", mock.Anything).Return(nil)
				oauthReadWriter.On("GetTokenInfo", userUUID).Return(nil, errors.New("not found"))
				oauthReadWriter.On("WriteTokenInfo", userUUID, mock.Anything).Return(nil)

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

				uri := fmt.Sprintf("%s?state=%s&code=%s", uriPrefix, challenge, "")
				req, rec := setupFakeEndpoint(method, uri, "")
				c := e.NewContext(req, rec)
				manager.V1OauthGoogleCallback(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				// Check the cookie exists
				_, err := utils.GetCookieByName(rec.Result().Cookies(), "x-authentication-bearer")
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})

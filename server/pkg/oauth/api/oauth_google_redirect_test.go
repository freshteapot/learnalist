package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	oauthApi "github.com/freshteapot/learnalist-api/server/pkg/oauth/api"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"golang.org/x/oauth2"
)

var _ = Describe("Testing Google Oauth redirect", func() {

	var (
		logger *logrus.Logger

		method    string
		uriPrefix string
		uri       string

		c   echo.Context
		e   *echo.Echo
		req *http.Request
		rec *httptest.ResponseRecorder

		hugoHelper     *mocks.HugoSiteBuilder
		userManagement *mocks.Management
		userSession    *mocks.Session
		userFromIDP    *mocks.UserFromIDP
		oauthHandlers  oauth.Handlers
		oauth2Config   *mocks.OAuth2ConfigInterface
		service        oauthApi.OauthService

		challenge string
		want      error
	)

	BeforeEach(func() {
		logger, _ = test.NewNullLogger()

		userFromIDP = &mocks.UserFromIDP{}
		userManagement = &mocks.Management{}
		userSession = &mocks.Session{}
		oauth2Config = &mocks.OAuth2ConfigInterface{}
		oauthHandlers = oauth.Handlers{}
		oauthHandlers.AddGoogle(oauth2Config)
		hugoHelper = &mocks.HugoSiteBuilder{}

		service = oauthApi.NewService(
			userManagement,
			hugoHelper,
			oauthHandlers,
			userSession,
			userFromIDP,
			logger)

		method = http.MethodGet
		uriPrefix = "/api/v1/oauth/google/redirect"
		uri = uriPrefix
		e = echo.New()

		challenge = "fake-challenge-123"
		want = errors.New("fail")
	})

	It("Not enabled on server", func() {
		oauthHandlers.Google = nil
		service = oauthApi.NewService(
			userManagement,
			hugoHelper,
			oauthHandlers,
			userSession,
			userFromIDP,
			logger)

		req, rec = testutils.SetupJSONEndpoint(method, uri, "")
		c = e.NewContext(req, rec)
		c.SetPath(uri)

		service.V1OauthGoogleRedirect(c)
		Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`this website has not configured Google OAuth`))
	})

	It("Creating challenge fails due to repo issue", func() {
		req, rec = testutils.SetupJSONEndpoint(method, uri, "")
		c = e.NewContext(req, rec)
		c.SetPath(uri)
		userSession.On("CreateWithChallenge").Return("", want)

		service.V1OauthGoogleRedirect(c)
		Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(i18n.InternalServerErrorFunny))
	})

	It("Challenge created, return redirect", func() {
		req, rec = testutils.SetupJSONEndpoint(method, uri, "")
		c = e.NewContext(req, rec)
		c.SetPath(uri)
		fakeRedirectURL := "https://learnalist.net/fake-redirect"
		userSession.On("CreateWithChallenge").Return(challenge, nil)
		oauth2Config.On("AuthCodeURL", challenge, oauth2.AccessTypeOffline).Return(fakeRedirectURL)

		service.V1OauthGoogleRedirect(c)
		Expect(rec.Code).To(Equal(http.StatusFound))
		a, _ := rec.Result().Location()
		Expect(a.String()).To(Equal(fakeRedirectURL))
	})
})

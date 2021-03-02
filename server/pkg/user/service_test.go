package user_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing User from IDP", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder

		service user.UserService

		oauthHandlers oauth.Handlers
		userFromIDP   *mocks.UserFromIDP
		userSession   *mocks.Session
		hugoHelper    *mocks.HugoSiteBuilder

		want error

		userUUID string
		uri      string
	)

	BeforeEach(func() {
		logger, _ = test.NewNullLogger()
		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		eventMessageBus.On("Subscribe", event.TopicMonolog, "spacedRepetitionService", mock.Anything)
		e = echo.New()

		uri = "/api/v1/api/v1/user/login/idp"
		want = errors.New("want")
		userUUID = "fake-user-123"
		oauthHandlers = oauth.Handlers{}
		oauthHandlers.AddAppleID(&mocks.OAuth2ConfigInterface{})
		oauthHandlers.AddGoogle(&mocks.OAuth2ConfigInterface{})

		hugoHelper = &mocks.HugoSiteBuilder{}
		userSession = &mocks.Session{}
		userFromIDP = &mocks.UserFromIDP{}

		service = user.NewService(
			oauthHandlers,
			userFromIDP,
			userSession,
			hugoHelper,
			logger)

		fmt.Println(service, want, userUUID)
	})

	It("Bad json input", func() {
		req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, `{bad json}`)
		c = e.NewContext(req, rec)
		c.SetPath(uri)
		service.LoginViaIDP(c)
		Expect(rec.Code).To(Equal(http.StatusBadRequest))
		testutils.CheckMessageResponseFromResponseRecorder(rec, "Check the documentation")
	})

	It("Idp not enabled / supported", func() {
		input := openapi.HttpUserLoginIdpInput{
			Idp: "fake",
		}
		b, _ := json.Marshal(input)
		req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))

		c = e.NewContext(req, rec)
		c.SetPath(uri)
		service.LoginViaIDP(c)
		Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
		testutils.CheckMessageResponseFromResponseRecorder(rec, "Idp not supported: apple,google")
	})

	It("Defense code, if we add idp but do not add the logic", func() {

	})

	It("Failed to get userUUID from the idp", func() {

	})

	When("Looking up the user", func() {
		It("Issue talking to via the repo", func() {

		})

		When("User not found, register", func() {
			It("Issue talking to via the repo", func() {

			})

			It("New User registered", func() {

			})
		})

		It("Failed to create session", func() {

		})

		It("Session created", func() {

		})
	})
})

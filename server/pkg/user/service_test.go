package user_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
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
)

var _ = Describe("Testing User from IDP", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		hook            *test.Hook
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder

		service user.UserService

		oauthHandlers oauth.Handlers
		userFromIDP   *mocks.UserFromIDP
		userSession   *mocks.Session
		hugoHelper    *mocks.HugoSiteBuilder
		oauthApple    *mocks.OAuth2ConfigInterface
		oauthGoogle   *mocks.OAuth2ConfigInterface

		want error

		userUUID   string
		uri        string
		inputBytes []byte
		input      openapi.HttpUserLoginIdpInput
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		eventMessageBus.On("Subscribe", event.TopicMonolog, "spacedRepetitionService", mock.Anything)
		e = echo.New()

		uri = "/api/v1/api/v1/user/login/idp"
		want = errors.New("want")
		userUUID = "fake-user-123"
		oauthHandlers = oauth.Handlers{}
		oauthGoogle = &mocks.OAuth2ConfigInterface{}
		oauthApple = &mocks.OAuth2ConfigInterface{}
		oauthHandlers.AddAppleID(oauthApple)
		oauthHandlers.AddGoogle(oauthGoogle)

		hugoHelper = &mocks.HugoSiteBuilder{}
		userSession = &mocks.Session{}
		userFromIDP = &mocks.UserFromIDP{}

		service = user.NewService(
			oauthHandlers,
			userFromIDP,
			userSession,
			hugoHelper,
			logger)

		fmt.Println(userUUID)

		input = openapi.HttpUserLoginIdpInput{
			Idp:     oauth.IDPKeyGoogle,
			IdToken: "FAKE",
		}
		inputBytes, _ = json.Marshal(input)
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
		req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(inputBytes))

		c = e.NewContext(req, rec)
		c.SetPath(uri)
		oauthGoogle.On("GetUserUUIDFromIDP", input).Return("", want)
		service.LoginViaIDP(c)
		Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)

		Expect(hook.LastEntry().Data["event"]).To(Equal("idp-token-verification"))
		Expect(hook.LastEntry().Data["error"]).To(Equal(want))
		Expect(hook.LastEntry().Data["idp"]).To(Equal(input.Idp))
	})

	When("Looking up the user", func() {
		BeforeEach(func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(inputBytes))
			c = e.NewContext(req, rec)
			c.SetPath(uri)
		})

		It("Issue talking to via the repo", func() {
			oauthGoogle.On("GetUserUUIDFromIDP", input).Return(userUUID, nil)
			userFromIDP.On("Lookup", oauth.IDPKeyGoogle, user.IDPKindUserID, userUUID).Return(userUUID, want)

			service.LoginViaIDP(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)

			Expect(hook.LastEntry().Data["event"]).To(Equal("idp-lookup-user-info"))
			Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			Expect(hook.LastEntry().Data["idp"]).To(Equal(input.Idp))
		})

		When("User not found, register", func() {
			It("Issue talking to via the repo", func() {
				oauthGoogle.On("GetUserUUIDFromIDP", input).Return(userUUID, nil)
				userFromIDP.On("Lookup", oauth.IDPKeyGoogle, user.IDPKindUserID, userUUID).Return(userUUID, utils.ErrNotFound)
				userFromIDP.On("Register", input.Idp, user.IDPKindUserID, userUUID, []byte(``)).Return(userUUID, want)

				service.LoginViaIDP(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)

				Expect(hook.LastEntry().Data["event"]).To(Equal("idp-register-user"))
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				Expect(hook.LastEntry().Data["idp"]).To(Equal(input.Idp))
			})

			It("New User registered", func() {
				oauthGoogle.On("GetUserUUIDFromIDP", input).Return(userUUID, nil)
				userFromIDP.On("Lookup", oauth.IDPKeyGoogle, user.IDPKindUserID, userUUID).Return(userUUID, utils.ErrNotFound)
				userFromIDP.On("Register", input.Idp, user.IDPKindUserID, userUUID, []byte(``)).Return(userUUID, nil)

				expectedEvents := []string{}
				verify := func(args mock.Arguments) {
					moment := args[1].(event.Eventlog)
					expectedEvents = append(expectedEvents, moment.Kind)
				}
				eventMessageBus.On("Publish", event.TopicMonolog, mock.Anything).Times(2).Run(verify)

				//eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(entry event.Eventlog) bool {
				//	b, _ := json.Marshal(entry.Data)
				//	var moment event.EventUser
				//	json.Unmarshal(b, &moment)
				//	Expect(entry.Kind).To(Equal(event.ApiUserRegister))
				//	Expect(moment.UUID).To(Equal(userUUID))
				//	Expect(moment.Kind).To(Equal(event.KindUserRegisterIDPGoogle))
				//	return true
				//}))

				hugoHelper.On("WriteListsByUser", userUUID, []alist.ShortInfo{}).Return(nil)
				aSession := user.UserSession{
					Token:     "hi",
					UserUUID:  userUUID,
					Challenge: "",
				}
				userSession.On("NewSession", userUUID).Return(aSession, nil)

				service.LoginViaIDP(c)

				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(expectedEvents).To(Equal([]string{event.ApiUserRegister, event.ApiUserLogin}))
				response := api.HTTPLoginResponse{
					Token:    "hi",
					UserUUID: userUUID,
				}
				b, _ := json.Marshal(response)
				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(string(b)))
			})
		})

		It("Failed to create session", func() {

		})

		It("Session created", func() {

		})
	})
})

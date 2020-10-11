package api_test

import (
	"errors"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing user login endpoint", func() {
	var (
		endpoint                    = "/api/v1/user/login"
		datastore                   *mocks.Datastore
		userWithUsernameAndPassword *mocks.UserWithUsernameAndPassword
		userSession                 *mocks.Session
	)
	AfterEach(emptyDatabase)

	BeforeEach(func() {
		datastore = &mocks.Datastore{}
		userWithUsernameAndPassword = &mocks.UserWithUsernameAndPassword{}
		userSession = &mocks.Session{}
		m.Datastore = datastore
	})

	Context("POST'ing an invalid input", func() {
		It("Bad JSON", func() {
			input := ""
			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)
			m.V1PostLogin(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.ApiUserLoginError)
		})

		It("Invalid password", func() {
			input := `{"username":"iamusera", "password":"test1"}`
			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			m.V1PostLogin(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.ApiUserLoginError)
		})

		It("Invalid username", func() {
			inputs := []string{
				`{"username":"iamu@", "password":"test123"}`,
				`{"username":"iamu", "password":"test123"}`,
			}

			for _, input := range inputs {
				req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
				e := echo.New()
				c := e.NewContext(req, rec)
				m.V1PostLogin(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.ApiUserLoginError)
			}
		})
	})

	Context("Login with valid input", func() {
		var (
			session user.UserSession
			input   = `{"username":"iamusera", "password":"test123"}`
		)
		BeforeEach(func() {
			datastore = &mocks.Datastore{}
			userWithUsernameAndPassword = &mocks.UserWithUsernameAndPassword{}
			userSession = &mocks.Session{}
			m.Datastore = datastore

			session.Token = "fake-token"
			session.UserUUID = "fake-123"
		})

		It("Correct credentials", func() {
			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			datastore.On("UserWithUsernameAndPassword").Return(userWithUsernameAndPassword)
			userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(session.UserUUID, nil)

			datastore.On("UserSession").Return(userSession)
			userSession.On("NewSession", session.UserUUID).
				Return(session, nil)
			eventMessageBus := &mocks.EventlogPubSub{}
			eventMessageBus.On("Publish", mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Kind).To(Equal(event.ApiUserLogin))
				Expect(moment.Data.(event.EventUser).UUID).To(Equal(session.UserUUID))
				Expect(moment.Data.(event.EventUser).Kind).To(Equal(event.KindUserLoginUsername))
				return true
			}))
			event.SetBus(eventMessageBus)

			m.V1PostLogin(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"token":"fake-token","user_uuid":"fake-123"}`))
		})

		It("Wrong credentials", func() {
			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			datastore.On("UserWithUsernameAndPassword").Return(userWithUsernameAndPassword)
			userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return("", errors.New("fake error"))

			m.V1PostLogin(c)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.AclHttpAccessDeny)
		})

		It("Failed to create a user session", func() {
			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			datastore.On("UserWithUsernameAndPassword").Return(userWithUsernameAndPassword)
			userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(session.UserUUID, nil)
			datastore.On("UserSession").Return(userSession)
			userSession.On("NewSession", session.UserUUID).
				Return(session, errors.New("fake"))

			m.V1PostLogin(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})
	})

})

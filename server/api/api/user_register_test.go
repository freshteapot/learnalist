package api_test

import (
	"errors"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	mockModels "github.com/freshteapot/learnalist-api/server/api/models/mocks"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Register user endpoint", func() {
	AfterEach(emptyDatabase)

	When("/register", func() {
		var userUUID string
		var datastore *mockModels.Datastore
		BeforeEach(func() {
			datastore = &mockModels.Datastore{}
			m.Datastore = datastore
		})

		It("POST'ing an invalid input", func() {
			input := ""
			req, rec := setupFakeEndpoint(http.MethodGet, "/api/v1/register", input)
			e := echo.New()
			c := e.NewContext(req, rec)
			//c.Set("loggedInUser", *user)
			m.V1PostRegister(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Please refer to the documentation on user registration"}`))
		})

		Context("POST'ing an invalid input", func() {
			It("Invalid password", func() {
				user := &uuid.User{
					Uuid: userUUID,
				}
				input := `{"username":"iamusera", "password":"test1"}`
				req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/register", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				datastore.On("GetUserByCredentials", mock.Anything).Return(&uuid.User{}, errors.New("Fail"))

				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Please refer to the documentation on user registration"}`))
			})

			It("Invalid username", func() {
				user := &uuid.User{
					Uuid: userUUID,
				}
				inputs := []string{
					`{"username":"iamu@", "password":"test123"}`,
					`{"username":"iamu", "password":"test123"}`,
				}

				for _, input := range inputs {
					req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/register", input)
					e := echo.New()
					c := e.NewContext(req, rec)
					c.Set("loggedInUser", *user)
					datastore.On("GetUserByCredentials", mock.Anything).Return(&uuid.User{}, errors.New("Fail"))

					m.V1PostRegister(c)
					Expect(rec.Code).To(Equal(http.StatusBadRequest))
					Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Please refer to the documentation on user registration"}`))
				}
			})
		})

		Context("Registering a valid user", func() {
			It("New user", func() {
				user := &uuid.User{
					Uuid: "fake-123",
				}
				input := `{"username":"iamusera", "password":"test123"}`
				req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/register", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				datastore.On("InsertNewUser", mock.Anything).Return(user, nil)

				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusCreated))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"uuid":"fake-123","username":"iamusera"}`))
			})

			It("New user, database issue via saving user", func() {
				user := &uuid.User{
					Uuid: "fake-123",
				}
				input := `{"username":"iamusera", "password":"test123"}`
				req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/register", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				datastore.On("GetUserByCredentials", mock.Anything).Return(user, errors.New(i18n.DatabaseLookupNotFound))
				datastore.On("InsertNewUser", mock.Anything).Return(user, errors.New("Fake"))

				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Sadly, our service has taken a nap."}`))
			})

			It("New user, but already exists, database issue", func() {
				user := &uuid.User{
					Uuid: "fake-123",
				}
				input := `{"username":"iamusera", "password":"test123"}`
				req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/register", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				datastore.On("InsertNewUser", mock.Anything).Return(user, errors.New(i18n.UserInsertUsernameExists))
				datastore.On("GetUserByCredentials", mock.Anything).Return(user, errors.New("Fake"))

				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Sadly, our service has taken a nap."}`))
			})

			It("New user, but already exists", func() {
				user := &uuid.User{
					Uuid: "fake-123",
				}
				input := `{"username":"iamusera", "password":"test123"}`
				req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/register", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				datastore.On("InsertNewUser", mock.Anything).Return(user, errors.New(i18n.UserInsertUsernameExists))
				datastore.On("GetUserByCredentials", mock.Anything).Return(user, nil)

				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"uuid":"fake-123","username":"iamusera"}`))
			})
		})
	})

})

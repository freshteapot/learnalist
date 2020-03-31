package api_test

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/user"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Register user endpoint", func() {
	AfterEach(emptyDatabase)

	When("/register", func() {
		var (
			datastore                   *mocks.Datastore
			userWithUsernameAndPassword *mocks.UserWithUsernameAndPassword
			endpoint                    = "/api/v1/user/register"
		)

		BeforeEach(func() {
			datastore = &mocks.Datastore{}
			userWithUsernameAndPassword = &mocks.UserWithUsernameAndPassword{}
			m.Datastore = datastore
		})

		Context("POST'ing invalid input", func() {
			It("Bad JSON", func() {
				input := ""
				req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
				e := echo.New()
				c := e.NewContext(req, rec)
				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Please refer to the documentation on user registration"}`))
			})

			It("Invalid password", func() {
				input := `{"username":"iamusera", "password":"test1"}`
				req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
				e := echo.New()
				c := e.NewContext(req, rec)

				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Please refer to the documentation on user registration"}`))
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
					m.V1PostRegister(c)
					Expect(rec.Code).To(Equal(http.StatusBadRequest))
					Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Please refer to the documentation on user registration"}`))
				}
			})
		})

		Context("Registering a valid user", func() {
			var (
				userInfo user.UserInfoFromUsernameAndPassword
				input    = `{"username":"iamusera", "password":"test123"}`
			)
			BeforeEach(func() {
				userInfo = user.UserInfoFromUsernameAndPassword{
					UserUUID: "fake-123",
					Username: "iamusera",
					Hash:     "na",
				}
			})

			It("New user", func() {
				req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
				e := echo.New()
				c := e.NewContext(req, rec)
				datastore.On("UserWithUsernameAndPassword").Return(userWithUsernameAndPassword)
				userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return("", sql.ErrNoRows)

				userWithUsernameAndPassword.On("Register", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(userInfo, nil)

				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusCreated))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"uuid":"fake-123","username":"iamusera"}`))
			})

			It("New user, database issue via saving user", func() {
				req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
				e := echo.New()
				c := e.NewContext(req, rec)

				datastore.On("UserWithUsernameAndPassword").Return(userWithUsernameAndPassword)
				userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return("", sql.ErrNoRows)
				userWithUsernameAndPassword.On("Register", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(userInfo, errors.New("Fake"))

				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Sadly, our service has taken a nap."}`))
			})

			It("New user, but already exists", func() {
				req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
				e := echo.New()
				c := e.NewContext(req, rec)

				datastore.On("UserWithUsernameAndPassword").Return(userWithUsernameAndPassword)
				userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(userInfo.UserUUID, nil)

				m.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"uuid":"fake-123","username":"iamusera"}`))
			})
		})
	})

})

package api_test

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	mockModels "github.com/freshteapot/learnalist-api/server/api/models/mocks"
	mockUser "github.com/freshteapot/learnalist-api/server/pkg/user/mocks"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing user logout endpoint", func() {
	var (
		endpoint    = "/api/v1/user/logout"
		datastore   *mockModels.Datastore
		userSession *mockUser.Session
	)
	AfterEach(emptyDatabase)

	BeforeEach(func() {
		datastore = &mockModels.Datastore{}
		userSession = &mockUser.Session{}
		m.Datastore = datastore
	})

	Context("POST'ing an invalid input", func() {
		It("Bad JSON", func() {
			input := ""
			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)
			m.V1PostLogout(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			CheckMessageResponse(rec, i18n.ApiUserLogoutError)
		})

		It("Validating the input paths", func() {
			inputs := []string{
				`{
					"kind":"naughty",
					"user_uuid":"fake-123",
					"token":""
				}`,
				`{
					"kind":"token",
					"user_uuid":"fake-123",
					"token":""
				}`,
				`{
					"kind":"user",
					"user_uuid":"fake-123",
					"token":""
				}`,
				`{
					"kind":"user",
					"user_uuid":"",
					"token":"fake-token"
				}`,
				`{
					"kind":"user",
					"user_uuid":"fake-123",
					"token":""
				}`,
			}

			for _, input := range inputs {
				req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
				e := echo.New()
				c := e.NewContext(req, rec)
				m.V1PostLogout(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				CheckMessageResponse(rec, i18n.ApiUserLogoutError)
			}
		})
	})

	Context("Remove sessions", func() {
		var (
			userUUID = "fake-user-123"
			token    = "fake-token-123"
		)
		BeforeEach(func() {
			datastore = &mockModels.Datastore{}
			userSession = &mockUser.Session{}
			m.Datastore = datastore
		})

		It("Remove session by token credentials", func() {
			input := fmt.Sprintf(`{
				"kind":"token",
				"user_uuid":"%s",
				"token":"%s"
			}`, userUUID, token)

			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			datastore.On("UserSession").Return(userSession)
			userSession.On("GetUserUUIDByToken", token).
				Return(userUUID, nil)
			userSession.On("RemoveSessionForUser", userUUID, token).
				Return(nil)

			m.V1PostLogout(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			CheckMessageResponse(rec, "Session fake-token-123, is now logged out")
		})

		It("Remove all sessions for a user", func() {
			input := fmt.Sprintf(`{
				"kind":"user",
				"user_uuid":"%s",
				"token":"%s"
			}`, userUUID, token)

			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			datastore.On("UserSession").Return(userSession)
			userSession.On("GetUserUUIDByToken", token).
				Return(userUUID, nil)
			userSession.On("RemoveSessionsForUser", userUUID).
				Return(nil)

			m.V1PostLogout(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			CheckMessageResponse(rec, "All sessions have been logged out for user fake-user-123")
		})

		It("Token doesnt exist", func() {
			input := fmt.Sprintf(`{
				"kind":"user",
				"user_uuid":"%s",
				"token":"%s"
			}`, userUUID, token)

			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			datastore.On("UserSession").Return(userSession)
			userSession.On("GetUserUUIDByToken", token).
				Return("", sql.ErrNoRows)

			m.V1PostLogout(c)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			CheckMessageResponse(rec, i18n.AclHttpAccessDeny)
		})

		It("Token lookup failed due to the database possibly", func() {
			input := fmt.Sprintf(`{
				"kind":"user",
				"user_uuid":"%s",
				"token":"%s"
			}`, userUUID, token)

			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			datastore.On("UserSession").Return(userSession)
			userSession.On("GetUserUUIDByToken", token).
				Return("", errors.New("fake"))

			m.V1PostLogout(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			CheckMessageResponse(rec, i18n.InternalServerErrorFunny)
		})

		It("User linked to the token is not the one in the payload", func() {
			input := fmt.Sprintf(`{
				"kind":"user",
				"user_uuid":"%s",
				"token":"%s"
			}`, userUUID, token)

			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			datastore.On("UserSession").Return(userSession)
			userSession.On("GetUserUUIDByToken", token).
				Return("abc", nil)

			m.V1PostLogout(c)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			CheckMessageResponse(rec, i18n.AclHttpAccessDeny)
		})

		It("Database issued when removing the sessions", func() {
			input := fmt.Sprintf(`{
				"kind":"user",
				"user_uuid":"%s",
				"token":"%s"
			}`, userUUID, token)

			req, rec := setupFakeEndpoint(http.MethodPost, endpoint, input)
			e := echo.New()
			c := e.NewContext(req, rec)

			datastore.On("UserSession").Return(userSession)
			userSession.On("GetUserUUIDByToken", token).
				Return(userUUID, nil)
			userSession.On("RemoveSessionsForUser", userUUID).
				Return(errors.New("fake"))

			m.V1PostLogout(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			CheckMessageResponse(rec, i18n.InternalServerErrorFunny)
		})
	})

})

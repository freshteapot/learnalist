package api_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userApi "github.com/freshteapot/learnalist-api/server/pkg/user/api"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing User API", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		hook            *test.Hook
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder

		service userApi.UserService

		oauthHandlers               oauth.Handlers
		userManagement              *mocks.Management
		userFromIDP                 *mocks.UserFromIDP
		userSession                 *mocks.Session
		userInfoRepo                *mocks.UserInfoRepository
		userWithUsernameAndPassword *mocks.UserWithUsernameAndPassword
		oauthApple                  *mocks.OAuth2ConfigInterface
		oauthGoogle                 *mocks.OAuth2ConfigInterface
		aclRepo                     *mocks.Acl

		want error

		loggedInUser *uuid.User
		userUUID     string
		endpoint     string
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		eventMessageBus.On("Subscribe", event.TopicMonolog, "userService", mock.Anything)
		e = echo.New()

		want = errors.New("want")
		userUUID = "fake-user-123"
		loggedInUser = &uuid.User{
			Uuid: userUUID,
		}

		oauthHandlers = oauth.Handlers{}
		oauthGoogle = &mocks.OAuth2ConfigInterface{}
		oauthApple = &mocks.OAuth2ConfigInterface{}
		oauthHandlers.AddAppleID(oauthApple)
		oauthHandlers.AddGoogle(oauthGoogle)

		aclRepo = &mocks.Acl{}
		userManagement = &mocks.Management{}
		userSession = &mocks.Session{}
		userFromIDP = &mocks.UserFromIDP{}
		userInfoRepo = &mocks.UserInfoRepository{}
		userWithUsernameAndPassword = &mocks.UserWithUsernameAndPassword{}
		service = userApi.NewService(
			oauthHandlers,
			aclRepo,
			userManagement,
			userFromIDP,
			userSession,
			userWithUsernameAndPassword,
			userInfoRepo,
			"",
			logger)

	})

	// @oas.path /v1/user/login
	When("Login via username", func() {
		BeforeEach(func() {
			endpoint = "/api/v1/user/login"
		})

		Context("POST'ing an invalid input", func() {
			It("Bad JSON", func() {
				input := `{`
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)
				service.V1PostLogin(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.ApiUserLoginError)
			})

			It("Invalid password", func() {
				input := `{"username":"iamusera", "password":"test1"}`
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)
				service.V1PostLogin(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.ApiUserLoginError)
			})

			It("Invalid username", func() {
				inputs := []string{
					`{"username":"iamu@", "password":"test123"}`,
					`{"username":"iamu", "password":"test123"}`,
				}

				for _, input := range inputs {
					req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
					c = e.NewContext(req, rec)
					c.SetPath(endpoint)
					service.V1PostLogin(c)
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
				session.Token = "fake-token"
				session.UserUUID = "fake-123"
			})

			It("Correct credentials", func() {
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(session.UserUUID, nil)

				userSession.On("NewSession", session.UserUUID).
					Return(session, nil)
				eventMessageBus := &mocks.EventlogPubSub{}
				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.ApiUserLogin))
					Expect(moment.Data.(event.EventUser).UUID).To(Equal(session.UserUUID))
					Expect(moment.Data.(event.EventUser).Kind).To(Equal(event.KindUserLoginUsername))
					return true
				}))
				event.SetBus(eventMessageBus)

				service.V1PostLogin(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"token":"fake-token","user_uuid":"fake-123"}`))
			})

			It("Wrong credentials", func() {
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return("", errors.New("fake error"))

				service.V1PostLogin(c)
				Expect(rec.Code).To(Equal(http.StatusForbidden))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.AclHttpAccessDeny)
			})

			It("Failed to create a user session", func() {
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(session.UserUUID, nil)
				userSession.On("NewSession", session.UserUUID).
					Return(session, errors.New("fake"))

				service.V1PostLogin(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
			})
		})
	})

	// @oas.path /v1/user/logout
	When("Logout", func() {
		Context("POST'ing an invalid input", func() {
			It("Bad JSON", func() {
				input := `{`
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)
				service.V1PostLogout(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.ApiUserLogoutError)
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
					req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
					c = e.NewContext(req, rec)
					c.SetPath(endpoint)
					service.V1PostLogout(c)
					Expect(rec.Code).To(Equal(http.StatusBadRequest))
					testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.ApiUserLogoutError)
				}
			})
		})

		Context("Remove sessions", func() {
			var (
				userUUID = "fake-user-123"
				token    = "fake-token-123"
			)

			BeforeEach(func() {
				endpoint = "/api/v1/user/logout"
			})

			It("Remove session by token credentials", func() {
				input := fmt.Sprintf(`{
					"kind":"token",
					"user_uuid":"%s",
					"token":"%s"
				}`, userUUID, token)

				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userSession.On("GetUserUUIDByToken", token).
					Return(userUUID, nil)
				userSession.On("RemoveSessionForUser", userUUID, token).
					Return(nil)

				eventMessageBus := &mocks.EventlogPubSub{}
				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.ApiUserLogout))
					Expect(moment.Data.(event.EventUser).UUID).To(Equal(userUUID))
					Expect(moment.Data.(event.EventUser).Kind).To(Equal(event.KindUserLogoutSession))
					return true
				}))
				event.SetBus(eventMessageBus)

				service.V1PostLogout(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Session fake-token-123, is now logged out")
			})

			It("Remove all sessions for a user", func() {
				input := fmt.Sprintf(`{
					"kind":"user",
					"user_uuid":"%s",
					"token":"%s"
				}`, userUUID, token)

				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userSession.On("GetUserUUIDByToken", token).
					Return(userUUID, nil)
				userSession.On("RemoveSessionsForUser", userUUID).
					Return(nil)

				eventMessageBus := &mocks.EventlogPubSub{}
				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.ApiUserLogout))
					Expect(moment.Data.(event.EventUser).UUID).To(Equal(userUUID))
					Expect(moment.Data.(event.EventUser).Kind).To(Equal(event.KindUserLogoutSessions))
					return true
				}))
				event.SetBus(eventMessageBus)
				service.V1PostLogout(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "All sessions have been logged out for user fake-user-123")
			})

			It("Token doesnt exist", func() {
				input := fmt.Sprintf(`{
					"kind":"user",
					"user_uuid":"%s",
					"token":"%s"
				}`, userUUID, token)

				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userSession.On("GetUserUUIDByToken", token).
					Return("", sql.ErrNoRows)

				service.V1PostLogout(c)
				Expect(rec.Code).To(Equal(http.StatusForbidden))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.AclHttpAccessDeny)
			})

			It("Token lookup failed due to the database possibly", func() {
				input := fmt.Sprintf(`{
					"kind":"user",
					"user_uuid":"%s",
					"token":"%s"
				}`, userUUID, token)

				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userSession.On("GetUserUUIDByToken", token).
					Return("", errors.New("fake"))

				service.V1PostLogout(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
			})

			It("User linked to the token is not the one in the payload", func() {
				input := fmt.Sprintf(`{
					"kind":"user",
					"user_uuid":"%s",
					"token":"%s"
				}`, userUUID, token)

				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userSession.On("GetUserUUIDByToken", token).
					Return("abc", nil)

				service.V1PostLogout(c)
				Expect(rec.Code).To(Equal(http.StatusForbidden))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.AclHttpAccessDeny)
			})

			It("Database issued when removing the sessions", func() {
				input := fmt.Sprintf(`{
					"kind":"user",
					"user_uuid":"%s",
					"token":"%s"
				}`, userUUID, token)

				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userSession.On("GetUserUUIDByToken", token).
					Return(userUUID, nil)
				userSession.On("RemoveSessionsForUser", userUUID).
					Return(errors.New("fake"))

				service.V1PostLogout(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)

			})
		})
	})

	// @oas.path /v1/user/login/idp
	When("Login via idp", func() {
		var (
			inputBytes []byte
			input      openapi.HttpUserLoginIdpInput
		)

		BeforeEach(func() {
			endpoint = "/api/v1/api/v1/user/login/idp"
			input = openapi.HttpUserLoginIdpInput{
				Idp:     oauth.IDPKeyGoogle,
				IdToken: "FAKE",
			}
			inputBytes, _ = json.Marshal(input)
		})

		It("Bad json input", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, `{bad json}`)
			c = e.NewContext(req, rec)
			c.SetPath(endpoint)
			service.LoginViaIDP(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Check the documentation")
		})

		It("Idp not enabled / supported", func() {
			input := openapi.HttpUserLoginIdpInput{
				Idp: "fake",
			}
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, string(b))

			c = e.NewContext(req, rec)
			c.SetPath(endpoint)
			service.LoginViaIDP(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Idp not supported: apple,google")
		})

		It("Defense code, if we add idp but do not add the logic", func() {

		})

		It("Failed to get userUUID from the idp", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, string(inputBytes))

			c = e.NewContext(req, rec)
			c.SetPath(endpoint)
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
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, string(inputBytes))
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)
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

					aSession := user.UserSession{
						Token:     "hi",
						UserUUID:  userUUID,
						Challenge: "",
					}
					userSession.On("NewSession", userUUID).Return(aSession, nil)

					service.LoginViaIDP(c)

					Expect(rec.Code).To(Equal(http.StatusOK))
					Expect(expectedEvents).To(Equal([]string{event.ApiUserRegister, event.ApiUserLogin}))
					response := openapi.HttpUserLoginResponse{
						Token:    "hi",
						UserUuid: userUUID,
					}
					b, _ := json.Marshal(response)
					Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(string(b)))
				})
			})

			It("Failed to create session", func() {
				oauthGoogle.On("GetUserUUIDFromIDP", input).Return(userUUID, nil)
				userFromIDP.On("Lookup", oauth.IDPKeyGoogle, user.IDPKindUserID, userUUID).Return(userUUID, nil)
				userSession.On("NewSession", userUUID).Return(user.UserSession{}, want)

				service.LoginViaIDP(c)

				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
			})

			It("Session created", func() {
				input.Idp = oauth.IDPKeyApple
				inputBytes, _ = json.Marshal(input)
				// Override BeforeEach to set the new input
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, string(inputBytes))
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				oauthApple.On("GetUserUUIDFromIDP", input).Return(userUUID, nil)
				userFromIDP.On("Lookup", oauth.IDPKeyApple, user.IDPKindUserID, userUUID).Return(userUUID, utils.ErrNotFound)
				userFromIDP.On("Register", input.Idp, user.IDPKindUserID, userUUID, []byte(``)).Return(userUUID, nil)

				expectedEvents := []string{}
				verify := func(args mock.Arguments) {
					moment := args[1].(event.Eventlog)
					expectedEvents = append(expectedEvents, moment.Kind)
				}
				eventMessageBus.On("Publish", event.TopicMonolog, mock.Anything).Times(2).Run(verify)

				aSession := user.UserSession{
					Token:     "hi",
					UserUUID:  userUUID,
					Challenge: "",
				}
				userSession.On("NewSession", userUUID).Return(aSession, nil)

				service.LoginViaIDP(c)

				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(expectedEvents).To(Equal([]string{event.ApiUserRegister, event.ApiUserLogin}))
				response := openapi.HttpUserLoginResponse{
					Token:    "hi",
					UserUuid: userUUID,
				}
				b, _ := json.Marshal(response)
				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(string(b)))
			})
		})

		It("How to add acl info about able to write public lists", func() {

			prefs := user.UserPreference{}
			prefs.Acl = user.ACL{
				PublicListWrite: 0,
			}

			prefs.Acl.PublicListWrite = 1

			b, _ := json.Marshal(prefs)
			fmt.Println(string(b))
		})
	})

	// @oas.path /v1/user
	When("Deleting a user", func() {
		var (
			session  user.UserSession
			endpoint string
		)

		BeforeEach(func() {

			session.Token = "fake-token"
			session.UserUUID = userUUID
			endpoint = fmt.Sprintf("/api/v1/user/delete/%s", userUUID)

			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/alist/:uuid")
			c.Set("loggedInUser", *loggedInUser)
		})

		It("The user to delete is not the same as the user logged in", func() {
			c.SetParamNames("uuid")
			c.SetParamValues("fake-345")
			service.V1DeleteUser(c)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
		})

		When("User is deleting themselves", func() {
			It("Issue deleting from the system", func() {
				c.SetParamNames("uuid")
				c.SetParamValues(userUUID)
				want := errors.New("fail")
				userManagement.On("DeleteUser", userUUID).Return(want)
				service.V1DeleteUser(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Sadly, our service has taken a nap.")
			})

			It("Successfully deleted user", func() {
				c.SetParamNames("uuid")
				c.SetParamValues(userUUID)
				userManagement.On("DeleteUser", userUUID).Return(nil)
				eventMessageBus := &mocks.EventlogPubSub{}
				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.ApiUserDelete))
					Expect(moment.UUID).To(Equal(userUUID))
					return true
				}))
				event.SetBus(eventMessageBus)

				service.V1DeleteUser(c)

				Expect(rec.Code).To(Equal(http.StatusOK))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "User has been removed")
			})
		})
	})

	// @oas.path /v1/user/info
	When("Asking for user info", func() {
		When("Get User info entry", func() {
			BeforeEach(func() {
				uri := fmt.Sprintf("/api/v1/user/info/%s", userUUID)
				req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
				c = e.NewContext(req, rec)
				c.SetPath("/api/v1//user/info/:uuid")
				c.Set("loggedInUser", *loggedInUser)
				c.SetParamNames("uuid")
				c.SetParamValues(userUUID)
			})

			It("Only your user", func() {
				c.SetParamNames("uuid")
				c.SetParamValues("fake-user-456")

				service.V1GetUserInfo(c)
				Expect(rec.Code).To(Equal(http.StatusForbidden))
			})

			It("Failed to talk to repo", func() {
				userInfoRepo.On("Get", userUUID).Return(user.UserPreference{}, want)
				service.V1GetUserInfo(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
			})

			When("Repo returns", func() {
				It("No info found", func() {
					userInfoRepo.On("Get", userUUID).Return(user.UserPreference{}, nil)
					service.V1GetUserInfo(c)
					Expect(rec.Code).To(Equal(http.StatusOK))
					Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"user_uuid":"fake-user-123","acl":{"list_public_write":0}}`))
				})

				When("Daily reminder", func() {
					It("empty", func() {
						userInfoRepo.On("Get", userUUID).Return(user.UserPreference{}, nil)
						service.V1GetUserInfo(c)
						Expect(rec.Code).To(Equal(http.StatusOK))
						Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"user_uuid":"fake-user-123","acl":{"list_public_write":0}}`))
					})

					It("Contains remind", func() {
						userInfoRepo.On("Get", userUUID).Return(user.UserPreference{
							DailyReminder: &user.UserPreferenceDailyReminder{
								RemindV1: &openapi.RemindDailySettings{
									TimeOfDay:     "09:00",
									Tz:            "Europe/Oslo",
									AppIdentifier: "remind_v1",
									Medium:        []string{"email"},
								},
							},
						}, nil)

						service.V1GetUserInfo(c)
						Expect(rec.Code).To(Equal(http.StatusOK))
						Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"user_uuid":"fake-user-123","daily_reminder":{"remind_v1":{"time_of_day":"09:00","tz":"Europe/Oslo","app_identifier":"remind_v1","medium":["email"]}},"acl":{"list_public_write":0}}`))
					})

					It("Contains plank", func() {
						userInfoRepo.On("Get", userUUID).Return(user.UserPreference{
							DailyReminder: &user.UserPreferenceDailyReminder{
								PlankV1: &openapi.RemindDailySettings{
									TimeOfDay:     "09:00",
									Tz:            "Europe/Oslo",
									AppIdentifier: "plank_v1",
									Medium:        []string{"email"},
								},
							},
						}, nil)

						service.V1GetUserInfo(c)
						Expect(rec.Code).To(Equal(http.StatusOK))
						Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"user_uuid":"fake-user-123","daily_reminder":{"plank_v1":{"time_of_day":"09:00","tz":"Europe/Oslo","app_identifier":"plank_v1","medium":["email"]}},"acl":{"list_public_write":0}}`))
					})
				})
			})
		})
	})

	// @oas.path /user/register
	When("/register", func() {
		BeforeEach(func() {
			endpoint = "/api/v1/user/register"
		})

		Context("POST'ing invalid input", func() {
			It("Bad JSON", func() {
				input := ""
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				service.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Please refer to the documentation on user registration")
			})

			It("Invalid password", func() {
				input := `{"username":"iamusera", "password":"test1"}`
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				service.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Please refer to the documentation on user registration")
			})

			It("Invalid username", func() {
				inputs := []string{
					`{"username":"iamu@", "password":"test123"}`,
					`{"username":"iamu", "password":"test123"}`,
				}

				for _, input := range inputs {
					req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
					c = e.NewContext(req, rec)
					c.SetPath(endpoint)
					service.V1PostRegister(c)
					Expect(rec.Code).To(Equal(http.StatusBadRequest))
					testutils.CheckMessageResponseFromResponseRecorder(rec, "Please refer to the documentation on user registration")
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
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return("", sql.ErrNoRows)

				userWithUsernameAndPassword.On("Register", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(userInfo, nil)

				eventMessageBus := &mocks.EventlogPubSub{}
				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.ApiUserRegister))
					Expect(moment.Data.(event.EventNewUser).Kind).To(Equal(event.KindUserRegisterUsername))
					return true
				}))
				event.SetBus(eventMessageBus)

				service.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusCreated))
				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"uuid":"fake-123","username":"iamusera"}`))
				userWithUsernameAndPassword.AssertExpectations(GinkgoT())

			})

			It("New user, database issue via saving user", func() {
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return("", sql.ErrNoRows)
				userWithUsernameAndPassword.On("Register", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(userInfo, errors.New("Fake"))

				service.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Sadly, our service has taken a nap.")
			})

			It("New user, but already exists", func() {
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, endpoint, input)
				c = e.NewContext(req, rec)
				c.SetPath(endpoint)

				userWithUsernameAndPassword.On("Lookup", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(userInfo.UserUUID, nil)

				service.V1PostRegister(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"uuid":"fake-123","username":"iamusera"}`))
			})
		})
	})
})

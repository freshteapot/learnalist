package app_settings_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/app_settings"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
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

var _ = Describe("Testing AppSettings API", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		hook            *test.Hook
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder
		service         app_settings.AppSettingsService
		repo            *mocks.ManagementStorage
		loggedInUser    *uuid.User
		want            error

		userUUID string
	)

	BeforeEach(func() {
		want = errors.New("fail")
		loggedInUser = &uuid.User{
			Uuid: "fake-user-123",
		}
		userUUID = loggedInUser.Uuid
		eventMessageBus = &mocks.EventlogPubSub{}
		eventMessageBus.On("Subscribe", event.TopicMonolog, "appSettingsService", mock.Anything)
		event.SetBus(eventMessageBus)

		e = echo.New()

		logger, hook = test.NewNullLogger()
		repo = &mocks.ManagementStorage{}
		service = app_settings.NewService(repo, logger)
	})

	When("Saving remind_v1 app settings", func() {
		It("Not valid json", func() {
			uri := "/api/v1/app-settings/remind_v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodPut, uri, `Hello`)
			c = e.NewContext(req, rec)
			c.Set("loggedInUser", *loggedInUser)

			service.SaveRemindV1(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Look at the documentation for more help")
		})

		It("Out of range PushEnabled", func() {
			pref := user.UserPreference{
				Apps: &user.UserPreferenceApps{
					RemindV1: &openapi.AppSettingsRemindV1{
						SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
							PushEnabled: 2,
						},
					},
				},
			}

			b, _ := json.Marshal(pref.Apps.RemindV1)
			rawJSON := string(b)
			uri := "/api/v1/app-settings/remind_v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodPut, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.Set("loggedInUser", *loggedInUser)

			service.SaveRemindV1(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "push_enabled can only be 1 or 0")
		})

		It("Issue talking to repo", func() {
			pref := user.UserPreference{
				Apps: &user.UserPreferenceApps{
					RemindV1: &openapi.AppSettingsRemindV1{
						SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
							PushEnabled: 1,
						},
					},
				},
			}

			b, _ := json.Marshal(pref.Apps.RemindV1)
			rawJSON := string(b)
			uri := "/api/v1/app-settings/remind_v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodPut, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.Set("loggedInUser", *loggedInUser)

			repo.On("GetInfo", userUUID).Return(nil, want)

			service.SaveRemindV1(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Dont save as already in the system", func() {
			pref := user.UserPreference{
				Apps: &user.UserPreferenceApps{
					RemindV1: &openapi.AppSettingsRemindV1{
						SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
							PushEnabled: 1,
						},
					},
				},
			}

			b, _ := json.Marshal(pref.Apps.RemindV1)
			rawJSON := string(b)
			uri := "/api/v1/app-settings/remind_v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodPut, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.Set("loggedInUser", *loggedInUser)

			prefB, _ := json.Marshal(pref)
			repo.On("GetInfo", userUUID).Return(prefB, nil)

			service.SaveRemindV1(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(rawJSON))
		})

		It("Saving, something went wrong talking to the repo", func() {
			pref := user.UserPreference{
				Apps: &user.UserPreferenceApps{
					RemindV1: &openapi.AppSettingsRemindV1{
						SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
							PushEnabled: 0,
						},
					},
				},
			}

			b, _ := json.Marshal(pref.Apps.RemindV1)
			rawJSON := string(b)
			uri := "/api/v1/app-settings/remind_v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodPut, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.Set("loggedInUser", *loggedInUser)

			repo.On("GetInfo", userUUID).Return(nil, utils.ErrNotFound)

			prefB, _ := json.Marshal(pref)
			repo.On("SaveInfo", userUUID, prefB).Return(want)
			service.SaveRemindV1(c)

			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)

			Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			Expect(hook.LastEntry().Data["method"]).To(Equal("SaveRemindV1"))
		})

		It("Success when pref is empty", func() {
			pref := user.UserPreference{
				Apps: &user.UserPreferenceApps{
					RemindV1: &openapi.AppSettingsRemindV1{
						SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
							PushEnabled: 0,
						},
					},
				},
			}

			b, _ := json.Marshal(pref.Apps.RemindV1)
			rawJSON := string(b)
			uri := "/api/v1/app-settings/remind_v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodPut, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.Set("loggedInUser", *loggedInUser)

			repo.On("GetInfo", userUUID).Return([]byte(`{}`), nil)

			prefB, _ := json.Marshal(pref)
			repo.On("SaveInfo", userUUID, prefB).Return(nil)

			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Kind).To(Equal(event.ApiAppSettingsRemindV1))
				Expect(moment.Data.(openapi.AppSettingsRemindV1).SpacedRepetition.PushEnabled).To(Equal(int32(0)))
				return true
			}))

			service.SaveRemindV1(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(rawJSON))

		})

		It("Success", func() {
			pref := user.UserPreference{
				Apps: &user.UserPreferenceApps{
					RemindV1: &openapi.AppSettingsRemindV1{
						SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
							PushEnabled: 0,
						},
					},
				},
			}

			b, _ := json.Marshal(pref.Apps.RemindV1)
			rawJSON := string(b)
			uri := "/api/v1/app-settings/remind_v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodPut, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.Set("loggedInUser", *loggedInUser)

			repo.On("GetInfo", userUUID).Return(nil, utils.ErrNotFound)

			prefB, _ := json.Marshal(pref)
			repo.On("SaveInfo", userUUID, prefB).Return(nil)

			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Kind).To(Equal(event.ApiAppSettingsRemindV1))
				Expect(moment.Data.(openapi.AppSettingsRemindV1).SpacedRepetition.PushEnabled).To(Equal(int32(0)))
				return true
			}))

			service.SaveRemindV1(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(rawJSON))
		})
	})

})

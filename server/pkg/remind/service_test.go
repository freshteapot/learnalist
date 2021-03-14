package remind_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

func TimeIn(t time.Time, name string) (time.Time, error) {

	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}
	return t, err
}

var _ = Describe("Testing Remind API", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder
		service         remind.RemindService
		repo            *mocks.UserInfoRepository
		loggedInUser    *uuid.User
	)

	BeforeEach(func() {
		loggedInUser = &uuid.User{
			Uuid: "fake-123",
		}
		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)

		e = echo.New()

		logger, _ = test.NewNullLogger()
		repo = &mocks.UserInfoRepository{}
		service = remind.NewService(repo, logger)

	})

	It("timezone fun", func() {
		/*
			- Load stream

			- Get list of timezones - from db
			- Loop over check map for time - from db
			- If found lookup users
			- Has record? yes OR no.
			- Send message
				- Will need token
				- Will need display_name
			- Send record back into topic, to make sure its still in the system
		*/
		for _, name := range []string{
			"",
			"Local",
			"Europe/Oslo",
			"Asia/Shanghai",
			"America/Metropolis",
		} {
			t, err := TimeIn(time.Now(), name)
			if err == nil {
				fmt.Println(t.Location(), t.Format("15:04"))
			} else {
				fmt.Println(name, "<time unknown>")
			}
		}
	})

	It("Validate time_of_day", func() {
		want := errors.New("fail")
		tests := []struct {
			input  string
			expect error
		}{
			{
				input:  "00:00:00",
				expect: want,
			},
			{
				input:  "car:00",
				expect: want,
			},
			{
				input:  "00:car",
				expect: want,
			},
			{
				input:  "00:00",
				expect: nil,
			},
			{
				input:  "000:00",
				expect: want,
			},
			{
				input:  "00:000",
				expect: want,
			},
			{
				input:  "25:00",
				expect: want,
			},
			{
				input:  "-1:60", // under 0
				expect: want,
			},
			{
				input:  "00:60", // under 0
				expect: want,
			},
			{
				input:  "00:-1", // under 0
				expect: want,
			},
		}

		for _, test := range tests {
			err := remind.ValidateTimeOfDay(test.input)
			if test.expect != nil {
				Expect(err).To(Equal(test.expect))
				continue
			}
			Expect(err).To(BeNil())
		}
	})

	When("Getting daily settings", func() {

		It("Not valid app", func() {
			appIdentifier := "remind:v1"
			uri := "/api/v1/remind/daily/remind:v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/remind/daily/:appIdentifier")
			c.SetParamNames("appIdentifier")
			c.SetParamValues(appIdentifier)
			c.Set("loggedInUser", *loggedInUser)
			service.GetDailySettings(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "appIdentifier is not valid")
		})

		It("Repo error getting data", func() {
			want := errors.New("fail")
			appIdentifier := "remind_v1"
			uri := "/api/v1/remind/daily/remind_v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/remind/daily/:appIdentifier")
			c.SetParamNames("appIdentifier")
			c.SetParamValues(appIdentifier)
			c.Set("loggedInUser", *loggedInUser)
			repo.On("Get", "fake-123").Return(user.UserPreference{}, want)
			service.GetDailySettings(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)

		})

		It("Settings not found", func() {
			appIdentifier := "remind_v1"
			uri := "/api/v1/remind/daily/remind_v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/remind/daily/:appIdentifier")
			c.SetParamNames("appIdentifier")
			c.SetParamValues(appIdentifier)
			c.Set("loggedInUser", *loggedInUser)
			repo.On("Get", "fake-123").Return(user.UserPreference{}, nil)
			service.GetDailySettings(c)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Settings not found")
		})

		It("Remind V1 app found", func() {
			appIdentifier := "remind_v1"
			uri := "/api/v1/remind/daily/remind_v1"

			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/remind/daily/:appIdentifier")
			c.SetParamNames("appIdentifier")
			c.SetParamValues(appIdentifier)
			c.Set("loggedInUser", *loggedInUser)

			repo.On("Get", "fake-123").Return(user.UserPreference{
				DailyReminder: &user.UserPreferenceDailyReminder{
					RemindV1: &openapi.RemindDailySettings{
						AppIdentifier: apps.RemindV1,
						TimeOfDay:     "12:16",
						Tz:            "Europe/Oslo",
						Medium:        []string{"push"},
					},
				},
			}, nil)
			service.GetDailySettings(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			want := `{"time_of_day":"12:16","tz":"Europe/Oslo","app_identifier":"remind_v1","medium":["push"]}`
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(want))
		})

		It("Plank V1 app found", func() {
			appIdentifier := "plank_v1"
			uri := "/api/v1/remind/daily/plank_v1"

			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/remind/daily/:appIdentifier")
			c.SetParamNames("appIdentifier")
			c.SetParamValues(appIdentifier)
			c.Set("loggedInUser", *loggedInUser)

			repo.On("Get", "fake-123").Return(user.UserPreference{
				DailyReminder: &user.UserPreferenceDailyReminder{
					PlankV1: &openapi.RemindDailySettings{
						AppIdentifier: apps.PlankV1,
						TimeOfDay:     "12:16",
						Tz:            "Europe/Oslo",
						Medium:        []string{"push"},
					},
				},
			}, nil)
			service.GetDailySettings(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			want := `{"time_of_day":"12:16","tz":"Europe/Oslo","app_identifier":"plank_v1","medium":["push"]}`
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(want))
		})
	})

	When("Deleting daily settings", func() {
		var (
			appIdentifier   = "plank_v1"
			uri             = "/api/v1/remind/daily/plank_v1"
			userPreferences user.UserPreference
		)

		BeforeEach(func() {
			userPreferences = user.UserPreference{
				DailyReminder: &user.UserPreferenceDailyReminder{
					PlankV1: &openapi.RemindDailySettings{
						AppIdentifier: apps.PlankV1,
						TimeOfDay:     "12:16",
						Tz:            "Europe/Oslo",
						Medium:        []string{"push"},
					},
				},
			}
		})
		It("Not valid app", func() {
			appIdentifier := "remind:v1"
			uri := "/api/v1/remind/daily/remind:v1"
			req, rec = testutils.SetupJSONEndpoint(http.MethodDelete, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/remind/daily/:appIdentifier")
			c.SetParamNames("appIdentifier")
			c.SetParamValues(appIdentifier)
			c.Set("loggedInUser", *loggedInUser)
			service.DeleteDailySettings(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "appIdentifier is not valid")
		})

		It("Repo error getting data", func() {
			want := errors.New("fail")
			req, rec = testutils.SetupJSONEndpoint(http.MethodDelete, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/remind/daily/:appIdentifier")
			c.SetParamNames("appIdentifier")
			c.SetParamValues(appIdentifier)
			c.Set("loggedInUser", *loggedInUser)
			repo.On("Get", "fake-123").Return(user.UserPreference{}, want)
			service.DeleteDailySettings(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Delete failed due to repo issues on delete", func() {
			want := errors.New("fail")
			req, rec = testutils.SetupJSONEndpoint(http.MethodDelete, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/remind/daily/:appIdentifier")
			c.SetParamNames("appIdentifier")
			c.SetParamValues(appIdentifier)
			c.Set("loggedInUser", *loggedInUser)

			repo.On("Get", "fake-123").Return(userPreferences, nil)

			repo.On("Save", "fake-123", user.UserPreference{
				DailyReminder: &user.UserPreferenceDailyReminder{},
			},
			).Return(want)

			service.DeleteDailySettings(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Delete was succesful", func() {

			req, rec = testutils.SetupJSONEndpoint(http.MethodDelete, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/remind/daily/:appIdentifier")
			c.SetParamNames("appIdentifier")
			c.SetParamValues(appIdentifier)
			c.Set("loggedInUser", *loggedInUser)

			repo.On("Get", "fake-123").Return(userPreferences, nil)
			repo.On("Save", "fake-123", user.UserPreference{
				DailyReminder: &user.UserPreferenceDailyReminder{},
			},
			).Return(nil)

			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Kind).To(Equal(remind.EventApiRemindDailySettings))
				Expect(moment.Action).To(Equal(event.ActionDeleted))
				Expect(moment.UUID).To(Equal(loggedInUser.Uuid))
				b, _ := json.Marshal(moment.Data)
				var settings openapi.RemindDailySettings
				json.Unmarshal(b, &settings)

				Expect(settings.AppIdentifier).To(Equal(apps.PlankV1))
				Expect(settings.TimeOfDay).To(Equal("12:16"))
				Expect(settings.Tz).To(Equal("Europe/Oslo"))
				Expect(settings.Medium[0]).To(Equal("push"))
				return true
			}))

			service.DeleteDailySettings(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})
})

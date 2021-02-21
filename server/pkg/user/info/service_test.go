package info_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user/info"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing User Info Service API", func() {
	var (
		logger          *logrus.Logger
		eventMessageBus *mocks.EventlogPubSub

		c   echo.Context
		e   *echo.Echo
		req *http.Request
		rec *httptest.ResponseRecorder

		service            info.UserInfoService
		userManagementRepo *mocks.ManagementStorage

		loggedInUser *uuid.User
		want         error

		userUUID string
	)

	BeforeEach(func() {
		want = errors.New("want")
		loggedInUser = &uuid.User{
			Uuid: "fake-user-123",
		}

		userUUID = loggedInUser.Uuid

		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		e = echo.New()

		logger, _ = test.NewNullLogger()
		userManagementRepo = &mocks.ManagementStorage{}
		eventMessageBus.On("Subscribe", event.TopicMonolog, "userInfoService", mock.Anything)

		service = info.NewService(userManagementRepo, logger)
	})

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
			userManagementRepo.On("GetInfo", userUUID).Return([]byte(``), want)
			service.V1GetUserInfo(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		When("Repo returns", func() {
			It("No info found", func() {
				userManagementRepo.On("GetInfo", userUUID).Return([]byte(``), nil)
				service.V1GetUserInfo(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"user_uuid":"fake-user-123"}`))
			})

			When("Daily reminder", func() {
				It("empty", func() {
					infoInDB := `{"daily_reminder":{}}`
					userManagementRepo.On("GetInfo", userUUID).Return([]byte(infoInDB), nil)
					service.V1GetUserInfo(c)
					Expect(rec.Code).To(Equal(http.StatusOK))
					Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"user_uuid":"fake-user-123"}`))
				})

				It("Contains remind", func() {
					infoInDB := `{"daily_reminder":{"remind_v1":{"time_of_day":"09:00","tz":"Europe/Oslo","app_identifier":"remind_v1","medium":["email"]}}}`
					userManagementRepo.On("GetInfo", userUUID).Return([]byte(infoInDB), nil)
					service.V1GetUserInfo(c)
					Expect(rec.Code).To(Equal(http.StatusOK))
					Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"user_uuid":"fake-user-123","daily_reminder":{"remind_v1":{"time_of_day":"09:00","tz":"Europe/Oslo","app_identifier":"remind_v1","medium":["email"]}}}`))
				})

				It("Contains plank", func() {
					infoInDB := `{"daily_reminder":{"plank_v1":{"time_of_day":"09:00","tz":"Europe/Oslo","app_identifier":"plank_v1","medium":["email"]}}}`
					userManagementRepo.On("GetInfo", userUUID).Return([]byte(infoInDB), nil)
					service.V1GetUserInfo(c)
					Expect(rec.Code).To(Equal(http.StatusOK))
					Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`{"user_uuid":"fake-user-123","daily_reminder":{"plank_v1":{"time_of_day":"09:00","tz":"Europe/Oslo","app_identifier":"plank_v1","medium":["email"]}}}`))
				})
			})
		})
	})
})

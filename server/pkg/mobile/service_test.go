package mobile_test

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing API", func() {
	var (
		service         mobile.MobileService
		repo            *mocks.MobileRepository
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder
		user1           *uuid.User
		user2           *uuid.User
		user1AppRemind  openapi.MobileDeviceInfo
		user1AppPlank   openapi.MobileDeviceInfo
		token           string

		remindV1RawJSON = `
				{
					"token": "fake-token-123",
					"app_identifier": "remind_v1"
				}`
	)

	BeforeEach(func() {
		token = "fake-token-123"
		user1 = &uuid.User{
			Uuid: "fake-user-123",
		}
		user2 = &uuid.User{
			Uuid: "fake-user-456",
		}

		user1AppRemind = openapi.MobileDeviceInfo{
			Token:         token,
			AppIdentifier: apps.RemindV1,
			UserUuid:      user1.Uuid,
		}
		user1AppPlank = openapi.MobileDeviceInfo{
			Token:         token,
			AppIdentifier: apps.PlankV1,
			UserUuid:      user1.Uuid,
		}

		eventMessageBus = &mocks.EventlogPubSub{}
		eventMessageBus.On("Subscribe", event.TopicMonolog, "mobile", mock.Anything)
		event.SetBus(eventMessageBus)

		e = echo.New()

		logger, _ = test.NewNullLogger()
		repo = &mocks.MobileRepository{}
		service = mobile.NewService(repo, logger)
	})

	When("Registering mobile device", func() {
		var (
			uri    = "/api/v1/mobile/register-device"
			method = http.MethodPost
		)

		It("Missing token", func() {
			rawJSON := `
{
	"token": "",
	"app_identifier": "remind_v1"
}`

			req, rec = testutils.SetupJSONEndpoint(method, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/mobile/register-device")
			c.Set("loggedInUser", *user1)
			service.RegisterDevice(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Token cant be empty")
		})

		It("Not supported app", func() {
			rawJSON := `
{
	"token": "fake-token-123",
	"app_identifier": "remind:v1"
}`

			req, rec = testutils.SetupJSONEndpoint(method, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/mobile/register-device")
			c.Set("loggedInUser", *user1)
			service.RegisterDevice(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "App identifier is not supported: plank_v1,remind_v1")
		})

		It("If app == plank:v1", func() {
			rawJSON := `
{
	"token": "fake-token-123",
	"app_identifier": "plank:v1"
}`

			req, rec = testutils.SetupJSONEndpoint(method, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/mobile/register-device")
			c.Set("loggedInUser", *user1)
			repo.On("GetDevicesInfoByToken", "fake-token-123").Return([]openapi.MobileDeviceInfo{}, nil)
			repo.On("SaveDeviceInfo", user1AppPlank).Return(http.StatusCreated, nil)
			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				device := moment.Data.(openapi.MobileDeviceInfo)
				Expect(device.UserUuid).To(Equal(user1.Uuid))
				Expect(device.AppIdentifier).To(Equal(apps.PlankV1))
				Expect(device.Token).To(Equal("fake-token-123"))
				return true
			}))

			service.RegisterDevice(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Device registered")
		})

		It("Empty app Identifier set to plank_v1", func() {
			rawJSON := `
{
	"token": "fake-token-123",
	"app_identifier": ""
}`

			req, rec = testutils.SetupJSONEndpoint(method, uri, rawJSON)
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/mobile/register-device")
			c.Set("loggedInUser", *user1)
			repo.On("GetDevicesInfoByToken", "fake-token-123").Return([]openapi.MobileDeviceInfo{}, nil)
			repo.On("SaveDeviceInfo", user1AppPlank).Return(http.StatusCreated, nil)
			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				device := moment.Data.(openapi.MobileDeviceInfo)
				Expect(device.UserUuid).To(Equal(user1.Uuid))
				Expect(device.AppIdentifier).To(Equal(apps.PlankV1))
				Expect(device.Token).To(Equal("fake-token-123"))
				return true
			}))

			service.RegisterDevice(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Device registered")
		})

		It("Repo fails on looking up devices", func() {
			want := errors.New("fail")
			req, rec = testutils.SetupJSONEndpoint(method, uri, remindV1RawJSON)
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/mobile/register-device")
			c.Set("loggedInUser", *user1)
			repo.On("GetDevicesInfoByToken", "fake-token-123").Return([]openapi.MobileDeviceInfo{}, want)
			service.RegisterDevice(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Repo fails on saving device", func() {
			want := errors.New("fail")
			req, rec = testutils.SetupJSONEndpoint(method, uri, remindV1RawJSON)
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/mobile/register-device")
			c.Set("loggedInUser", *user1)
			repo.On("GetDevicesInfoByToken", "fake-token-123").Return([]openapi.MobileDeviceInfo{}, nil)
			repo.On("SaveDeviceInfo", user1AppRemind).Return(http.StatusInternalServerError, want)

			service.RegisterDevice(c)

			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		When("Token already registered", func() {
			It("Delete existing entry with appIdentifier and token", func() {
				user2Device := openapi.MobileDeviceInfo{
					Token:         "fake-token-123",
					AppIdentifier: apps.RemindV1,
					UserUuid:      user2.Uuid,
				}

				req, rec = testutils.SetupJSONEndpoint(method, uri, remindV1RawJSON)
				c = e.NewContext(req, rec)
				c.SetPath("/api/v1/mobile/register-device")
				c.Set("loggedInUser", *user1)
				repo.On("GetDevicesInfoByToken", "fake-token-123").Return([]openapi.MobileDeviceInfo{user2Device}, nil)
				repo.On("DeleteByApp", user2.Uuid, apps.RemindV1).Return(nil)

				repo.On("SaveDeviceInfo", user1AppRemind).Return(http.StatusCreated, nil)

				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					switch moment.Kind {
					case event.MobileDeviceRemove:
						Expect(moment.Kind).To(Equal(event.MobileDeviceRemove))
					case event.MobileDeviceRegistered:
						Expect(moment.Kind).To(Equal(event.MobileDeviceRegistered))
						device := moment.Data.(openapi.MobileDeviceInfo)
						Expect(device.UserUuid).To(Equal(user1.Uuid))
						Expect(device.AppIdentifier).To(Equal(apps.RemindV1))
						Expect(device.Token).To(Equal("fake-token-123"))
					default:
						Fail("Not supported kind")
					}
					return true
				}))

				service.RegisterDevice(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Device registered")
				eventMessageBus.AssertNumberOfCalls(GinkgoT(), "Publish", 2)
			})

			It("Entry exists for appIdentifier and token", func() {
				req, rec = testutils.SetupJSONEndpoint(method, uri, remindV1RawJSON)
				c = e.NewContext(req, rec)
				c.SetPath("/api/v1/mobile/register-device")
				c.Set("loggedInUser", *user1)
				repo.On("GetDevicesInfoByToken", "fake-token-123").Return([]openapi.MobileDeviceInfo{
					{
						Token:         "fake-token-123",
						AppIdentifier: apps.PlankV1,
						UserUuid:      user1.Uuid,
					},
					{
						Token:         "fake-token-123",
						AppIdentifier: apps.RemindV1,
						UserUuid:      user1.Uuid,
					},
				}, nil)
				repo.On("DeleteByApp", user2.Uuid, apps.RemindV1).Return(nil)
				repo.On("SaveDeviceInfo", user1AppRemind).Return(http.StatusCreated, nil)

				service.RegisterDevice(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Device registered")
			})
		})
	})
})

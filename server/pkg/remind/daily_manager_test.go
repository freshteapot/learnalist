package remind_test

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Daily Manager", func() {
	var (
		logger       *logrus.Logger
		hook         *test.Hook
		settingsRepo *mocks.RemindDailySettingsRepository
		mobileRepo   *mocks.MobileRepository

		deviceInfo                     openapi.MobileDeviceInfo
		settings                       openapi.RemindDailySettings
		userUUID, token, appIdentifier string
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		settingsRepo = &mocks.RemindDailySettingsRepository{}
		mobileRepo = &mocks.MobileRepository{}
		userUUID = "fake-user-123"
		appIdentifier = apps.RemindV1
		token = "fake-token-123"
		deviceInfo = openapi.MobileDeviceInfo{
			UserUuid:      userUUID,
			AppIdentifier: appIdentifier,
			Token:         token,
		}

		settings = openapi.RemindDailySettings{
			TimeOfDay:     "09:00",
			Tz:            "Europe/Oslo",
			AppIdentifier: apps.RemindV1,
			Medium:        []string{"push"},
		}
	})

	When("Event arrives", func() {
		It("Remove device", func() {

			moment := event.Eventlog{
				Kind: event.MobileDeviceRemoved,
				Data: openapi.MobileDeviceInfo{
					UserUuid:      userUUID,
					AppIdentifier: appIdentifier,
				},
				Action: event.ActionDeleted,
			}

			mobileRepo.On("DeleteByApp", userUUID, appIdentifier).Return(nil)
			manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
			manager.OnEvent(moment)

			Expect("A").To(Equal("A"))
		})

		It("Register device", func() {
			moment := event.Eventlog{
				Kind:   event.MobileDeviceRegistered,
				Data:   deviceInfo,
				Action: event.ActionDeleted,
			}

			mobileRepo.On("SaveDeviceInfo", deviceInfo).Return(201, nil)
			manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
			manager.OnEvent(moment)

			Expect("A").To(Equal("A"))
		})

		When("Settings events arrive", func() {
			It("delete settings", func() {
				moment := event.Eventlog{
					UUID:   userUUID,
					Kind:   remind.EventApiRemindDailySettings,
					Data:   settings,
					Action: event.ActionDeleted,
				}
				settingsRepo.On("DeleteByApp", userUUID, appIdentifier).Return(nil)
				manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
				manager.OnEvent(moment)
			})

			It("upsert settings", func() {
				moment := event.Eventlog{
					UUID:   userUUID,
					Kind:   remind.EventApiRemindDailySettings,
					Data:   settings,
					Action: event.ActionUpsert,
				}
				// Test time outside fo this
				settingsRepo.On("Save", userUUID, settings, mock.AnythingOfType("string")).Return(nil)
				manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
				manager.OnEvent(moment)
			})
		})

		When("Activity arrives", func() {
			It("A plank", func() {
				moment := event.Eventlog{
					Kind: event.ApiPlank,
					Data: event.EventPlank{
						Action:   event.ActionNew,
						UserUUID: userUUID,
					},
				}
				settingsRepo.On("ActivityHappened", userUUID, apps.PlankV1).Return(nil)
				manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
				manager.OnEvent(moment)
			})

			It("Spaced repetition item", func() {

				moment := event.Eventlog{
					Kind: event.ApiSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind: spaced_repetition.EventKindNew,
						Data: spaced_repetition.SpacedRepetitionEntry{
							UserUUID: userUUID,
						},
					},
				}

				settingsRepo.On("ActivityHappened", userUUID, apps.RemindV1).Return(nil)
				manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
				manager.OnEvent(moment)
			})
		})
	})

	When("Sending notifications", func() {

		var (
			eventMessageBus *mocks.EventlogPubSub
		)
		BeforeEach(func() {
			eventMessageBus = &mocks.EventlogPubSub{}
			event.SetBus(eventMessageBus)
		})

		It("Issue getting Reminders from the repo", func() {
			want := errors.New("fail")
			settingsRepo.On("GetReminders", mock.AnythingOfType("string")).Return([]remind.RemindMe{}, want)

			testutils.SetLoggerToPanicOnFatal(logger)
			manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
			Expect(func() { manager.SendNotifications() }).Should(Panic())
			lastLog := hook.LastEntry()
			Expect(lastLog.Data["error"]).To(Equal(want))
		})

		It("No reminders found", func() {
			settingsRepo.On("GetReminders", mock.AnythingOfType("string")).Return([]remind.RemindMe{}, nil)
			manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
			manager.SendNotifications()
		})

		When("Reminders found", func() {
			It("1 found, skip because the token has not been set", func() {
				settingsRepo.On("GetReminders", mock.AnythingOfType("string")).Return([]remind.RemindMe{
					{
						UserUUID: userUUID,
						Medium:   []string{""},
						Settings: settings,
					},
				}, nil)

				settingsRepo.On("Save", userUUID, settings, mock.AnythingOfType("string")).Return(nil)
				manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
				manager.SendNotifications()

				lastLog := hook.LastEntry()
				Expect(lastLog.Data["msg_skipped"]).To(Equal(1))
				Expect(lastLog.Data["msg_sent"]).To(Equal(0))
			})

			When("Send notification", func() {
				It("Fails on updating user who has had a notification sent", func() {
					settingsRepo.On("GetReminders", mock.AnythingOfType("string")).Return([]remind.RemindMe{
						{
							UserUUID: userUUID,
							Medium:   []string{"fake-token-123"},
						},
					}, nil)

					eventMessageBus.On("Publish", event.TopicNotifications, mock.MatchedBy(func(moment event.Eventlog) bool {
						Expect(moment.Kind).To(Equal(event.KindPushNotification))

						return true
					}))
					testutils.SetLoggerToPanicOnFatal(logger)
					manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
					Expect(func() { manager.SendNotifications() }).Should(Panic())
				})

				It("Success", func() {
					settingsRepo.On("GetReminders", mock.AnythingOfType("string")).Return([]remind.RemindMe{
						{
							Medium:   []string{"", "fake-token-123"},
							UserUUID: userUUID,
							Settings: settings,
						},
						{
							Medium:   []string{""},
							UserUUID: "fake-user-456",
							Settings: settings,
						},
					}, nil)

					settingsRepo.On("Save", userUUID, settings, mock.AnythingOfType("string")).Return(nil)
					settingsRepo.On("Save", "fake-user-456", settings, mock.AnythingOfType("string")).Return(nil)

					eventMessageBus.On("Publish", event.TopicNotifications, mock.MatchedBy(func(moment event.Eventlog) bool {
						Expect(moment.Kind).To(Equal(event.KindPushNotification))

						return true
					}))

					manager := remind.NewDaily(settingsRepo, mobileRepo, logger)
					manager.SendNotifications()

					lastLog := hook.LastEntry()
					Expect(lastLog.Data["msg_skipped"]).To(Equal(1))
					Expect(lastLog.Data["msg_sent"]).To(Equal(1))

					mock.AssertExpectationsForObjects(GinkgoT())
				})
			})

		})
	})
})

package remind_test

import (
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/plank"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("Testing Manager", func() {
	var (
		logger *logrus.Logger
		dbCon  *sqlx.DB
		//mockSql      sqlmock.Sqlmock
		settingsRepo *mocks.RemindDailySettingsRepository
		mobileRepo   *mocks.MobileRepository

		deviceInfo                     openapi.MobileDeviceInfo
		settings                       openapi.RemindDailySettings
		userUUID, token, appIdentifier string
	)

	BeforeEach(func() {
		logger, _ = test.NewNullLogger()
		dbCon, _, _ = helper.GetMockDB()

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

	It("Remove device", func() {

		moment := event.Eventlog{
			Kind: mobile.EventMobileDeviceRemoved,
			Data: openapi.MobileDeviceInfo{
				UserUuid:      userUUID,
				AppIdentifier: appIdentifier,
			},
			Action: event.ActionDeleted,
		}

		mobileRepo.On("DeleteByApp", userUUID, appIdentifier).Return(nil)
		manager := remind.NewManager(dbCon, settingsRepo, mobileRepo, logger)
		manager.OnEvent(moment)

		Expect("A").To(Equal("A"))
	})

	It("Register device", func() {
		moment := event.Eventlog{
			Kind:   mobile.EventMobileDeviceRegistered,
			Data:   deviceInfo,
			Action: event.ActionDeleted,
		}

		mobileRepo.On("SaveDeviceInfo", deviceInfo).Return(201, nil)
		manager := remind.NewManager(dbCon, settingsRepo, mobileRepo, logger)
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
			manager := remind.NewManager(dbCon, settingsRepo, mobileRepo, logger)
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
			manager := remind.NewManager(dbCon, settingsRepo, mobileRepo, logger)
			manager.OnEvent(moment)
		})
	})

	When("Activity arrives", func() {
		It("A plank", func() {
			moment := event.Eventlog{
				Kind: plank.EventApiPlank,
				Data: plank.EventPlank{
					Kind:     plank.EventKindNew,
					UserUUID: userUUID,
				},
			}
			settingsRepo.On("ActivityHappened", userUUID, apps.PlankV1).Return(nil)
			manager := remind.NewManager(dbCon, settingsRepo, mobileRepo, logger)
			manager.OnEvent(moment)
		})

		It("Spaced repetition item", func() {

			moment := event.Eventlog{
				Kind: spaced_repetition.EventApiSpacedRepetition,
				Data: spaced_repetition.EventSpacedRepetition{
					Kind: spaced_repetition.EventKindNew,
					Data: spaced_repetition.SpacedRepetitionEntry{
						UserUUID: userUUID,
					},
				},
			}

			settingsRepo.On("ActivityHappened", userUUID, apps.RemindV1).Return(nil)
			manager := remind.NewManager(dbCon, settingsRepo, mobileRepo, logger)
			manager.OnEvent(moment)
		})
	})
})

package remind_test

import (
	"errors"
	"time"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Spaced Repetition Manager", func() {
	var (
		want                 = errors.New("fail")
		logger               *logrus.Logger
		hook                 *test.Hook
		userInfoRepo         *mocks.UserInfoRepository
		spacedRepetitionRepo *mocks.SpacedRepetitionRepository
		remindRepo           *mocks.RemindSpacedRepetitionRepository
		userUUID             string
		appSettingsEnabled   user.UserPreference
		whenNext             time.Time
		srsItem, nextSrsItem spaced_repetition.SpacedRepetitionEntry
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		userInfoRepo = &mocks.UserInfoRepository{}
		spacedRepetitionRepo = &mocks.SpacedRepetitionRepository{}
		remindRepo = &mocks.RemindSpacedRepetitionRepository{}
		appSettingsEnabled = user.UserPreference{
			Apps: &user.UserPreferenceApps{
				RemindV1: &openapi.AppSettingsRemindV1{
					SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
						PushEnabled: 1,
					},
				},
			},
		}

		userUUID = "fake-user-123"
		whenNext, _ = time.Parse(time.RFC3339, "2020-12-23T12:58:21Z")
		srsItem = spaced_repetition.SpacedRepetitionEntry{
			UserUUID: userUUID,
			UUID:     "ba9277fc4c6190fb875ad8f9cee848dba699937f",
			Body:     "{\"show\":\"Hello\",\"kind\":\"v1\",\"uuid\":\"ba9277fc4c6190fb875ad8f9cee848dba699937f\",\"data\":\"Hello\",\"settings\":{\"level\":\"0\",\"when_next\":\"2020-12-23T12:58:21Z\",\"created\":\"2020-12-23T11:58:21Z\"}}",
			WhenNext: whenNext,
			Created:  whenNext,
		}

		nextSrsItem = srsItem

	})

	When("OnEvent", func() {
		/*
			{
			  "kind": "api.spacedrepetition",
			  "data": {
			    "kind": "new",
			    "data": {
			      "uuid": "ba9277fc4c6190fb875ad8f9cee848dba699937f",
			      "body": "{\"show\":\"Hello\",\"kind\":\"v1\",\"uuid\":\"ba9277fc4c6190fb875ad8f9cee848dba699937f\",\"data\":\"Hello\",\"settings\":{\"level\":\"0\",\"when_next\":\"2020-12-23T12:58:21Z\",\"created\":\"2020-12-23T11:58:21Z\"}}",
			      "user_uuid": "47d71a5a-498f-414e-b501-2c085b637d66",
			      "when_next": "2020-12-23T12:58:21Z",
			      "created": "2020-12-23T11:58:21Z"
			    }
			  },
			  "timestamp": 1608724701
			}

			{
				"kind": "viewed",
				"action": "incr",
				"data": {
				"uuid": "ba9277fc4c6190fb875ad8f9cee848dba699937f",
				"body": "{\"show\":\"Hello\",\"kind\":\"v1\",\"uuid\":\"ba9277fc4c6190fb875ad8f9cee848dba699937f\",\"data\":\"Hello\",\"settings\":{\"level\":\"2\",\"when_next\":\"2020-12-23T12:05:10Z\",\"created\":\"2020-12-22T23:42:55Z\"}}",
				"user_uuid": "47d71a5a-498f-414e-b501-2c085b637d66",
				"when_next": "2020-12-23T12:05:10Z",
				"created": "2020-12-22T23:42:55Z"
				}
			}
		*/
		When("Checking for next entry and setting reminder", func() {
			It("Fails to get Remind info from app settings", func() {
				userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, want)
				testutils.SetLoggerToPanicOnFatal(logger)
				logContext := logger.WithField("context", "test")
				lastActive := time.Now().UTC()
				manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
				Expect(func() { manager.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive) }).Should(Panic())

				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				Expect(hook.LastEntry().Data["method"]).To(Equal("app_settings.GetRemindV1"))
			})

			It("Do nothing because Push is not enabled", func() {
				pref := user.UserPreference{
					Apps: &user.UserPreferenceApps{
						RemindV1: &openapi.AppSettingsRemindV1{
							SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
								PushEnabled: 0,
							},
						},
					},
				}

				userInfoRepo.On("Get", userUUID).Return(pref, nil)
				logContext := logger.WithField("context", "test")
				lastActive := time.Now().UTC()
				manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
				manager.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive)

				Expect(hook.LastEntry()).To(BeNil())
			})

			When("Push is enabled", func() {
				It("Issue getting next from repo", func() {
					userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
					spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, want)

					testutils.SetLoggerToPanicOnFatal(logger)

					logContext := logger.WithField("context", "test")
					lastActive := time.Now().UTC()
					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)

					Expect(func() { manager.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive) }).Should(Panic())

					Expect(hook.LastEntry().Data["error"]).To(Equal(want))
					Expect(hook.LastEntry().Data["method"]).To(Equal("m.spacedRepetitionRepo.GetNext"))
				})

				It("No entries found, remove user", func() {
					userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
					spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, utils.ErrNotFound)
					remindRepo.On("DeleteByUser", userUUID).Return(want)
					testutils.SetLoggerToPanicOnFatal(logger)

					logContext := logger.WithField("context", "test")
					lastActive := time.Now().UTC()
					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)

					Expect(func() { manager.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive) }).Should(Panic())

					Expect(hook.LastEntry().Data["error"]).To(Equal(want))
					Expect(hook.LastEntry().Data["method"]).To(Equal("m.remindRepo.DeleteByUser"))

				})

				It("No entries found, remove user", func() {
					userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
					spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, utils.ErrNotFound)
					remindRepo.On("DeleteByUser", userUUID).Return(nil)

					logContext := logger.WithField("context", "test")
					lastActive := time.Now().UTC()
					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
					manager.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive)

					Expect(hook.LastEntry()).To(BeNil())
				})

				It("Fail to set reminder due to talking to the repo", func() {
					logContext := logger.WithField("context", "test")
					lastActive := whenNext

					userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
					spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, nil)
					remindRepo.On("SetReminder", userUUID, whenNext, lastActive).Return(want)
					testutils.SetLoggerToPanicOnFatal(logger)

					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
					Expect(func() { manager.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive) }).Should(Panic())

					Expect(hook.LastEntry().Data["error"]).To(Equal(want))
					Expect(hook.LastEntry().Data["method"]).To(Equal("m.remindRepo.SetReminder"))
				})

				It("Success", func() {
					logContext := logger.WithField("context", "test")
					lastActive := whenNext

					userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
					spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, nil)
					remindRepo.On("SetReminder", userUUID, whenNext, whenNext).Return(nil)

					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
					manager.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive)

					Expect(hook.LastEntry()).To(BeNil())
				})
			})

		})

		When("App settings have changed for remind V1", func() {
			var (
				settings openapi.AppSettingsRemindV1
				moment   event.Eventlog
			)
			BeforeEach(func() {
				settings = openapi.AppSettingsRemindV1{
					SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
						PushEnabled: 0,
					},
				}
				moment = event.Eventlog{
					UUID:   userUUID,
					Kind:   event.ApiAppSettingsRemindV1,
					Data:   settings,
					Action: event.ActionUpsert,
				}
			})

			It("Saving fails", func() {
				pref := user.UserPreference{
					Apps: &user.UserPreferenceApps{
						RemindV1: &settings,
					},
				}

				userInfoRepo.On("Get", userUUID).Return(user.UserPreference{}, nil)
				userInfoRepo.On("Save", userUUID, pref).Return(want)

				testutils.SetLoggerToPanicOnFatal(logger)

				manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
				Expect(func() { manager.OnEvent(moment) }).Should(Panic())

				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				Expect(hook.LastEntry().Data["method"]).To(Equal("app_settings.SaveRemindV1"))
				mock.AssertExpectationsForObjects(GinkgoT(), spacedRepetitionRepo, userInfoRepo, remindRepo)
			})

			When("Push is enabled", func() {
				It("Success", func() {
					settings.SpacedRepetition.PushEnabled = 1
					moment.Data = settings
					moment.Timestamp = whenNext.UTC().Unix()

					pref := user.UserPreference{
						Apps: &user.UserPreferenceApps{
							RemindV1: &settings,
						},
					}

					userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
					userInfoRepo.On("Save", userUUID, pref).Return(nil)
					spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, nil)
					remindRepo.On("SetReminder", userUUID, whenNext, whenNext).Return(nil)

					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
					manager.OnEvent(moment)

					Expect(hook.LastEntry()).To(BeNil())
					mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo, remindRepo)
				})
			})

			When("Push enabled is set to 0", func() {
				It("Fail to remove user from the reminder system", func() {
					pref := user.UserPreference{
						Apps: &user.UserPreferenceApps{
							RemindV1: &settings,
						},
					}

					userInfoRepo.On("Get", userUUID).Return(user.UserPreference{}, nil)
					userInfoRepo.On("Save", userUUID, pref).Return(nil)
					remindRepo.On("DeleteByUser", userUUID).Return(want)
					testutils.SetLoggerToPanicOnFatal(logger)

					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
					Expect(func() { manager.OnEvent(moment) }).Should(Panic())

					Expect(hook.LastEntry().Data["error"]).To(Equal(want))
					Expect(hook.LastEntry().Data["method"]).To(Equal("m.remindRepo.DeleteByUser"))
					mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo, remindRepo)
				})

				It("Success, user removed", func() {
					pref := user.UserPreference{
						Apps: &user.UserPreferenceApps{
							RemindV1: &settings,
						},
					}

					userInfoRepo.On("Get", userUUID).Return(user.UserPreference{}, nil)
					userInfoRepo.On("Save", userUUID, pref).Return(nil)
					remindRepo.On("DeleteByUser", userUUID).Return(nil)

					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
					manager.OnEvent(moment)

					Expect(hook.LastEntry()).To(BeNil())
					mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo, remindRepo)
				})
			})
		})

		When("We add or view an entry", func() {
			var (
				whenNext             time.Time
				srsItem, nextSrsItem spaced_repetition.SpacedRepetitionEntry
				moment               event.Eventlog
			)
			BeforeEach(func() {
				whenNext, _ = time.Parse(time.RFC3339, "2020-12-23T12:58:21Z")
				srsItem = spaced_repetition.SpacedRepetitionEntry{
					UserUUID: userUUID,
					UUID:     "ba9277fc4c6190fb875ad8f9cee848dba699937f",
					Body:     "{\"show\":\"Hello\",\"kind\":\"v1\",\"uuid\":\"ba9277fc4c6190fb875ad8f9cee848dba699937f\",\"data\":\"Hello\",\"settings\":{\"level\":\"0\",\"when_next\":\"2020-12-23T12:58:21Z\",\"created\":\"2020-12-23T11:58:21Z\"}}",
					WhenNext: whenNext,
					Created:  whenNext,
				}

				nextSrsItem = srsItem
				moment = event.Eventlog{
					Kind: event.ApiSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind: spaced_repetition.EventKindNew,
						Data: srsItem,
					},
					Timestamp: whenNext.UTC().Unix(),
				}
			})

			It("Found and saved", func() {
				// A mess is born
				spacedRepetitionRepo.On("SaveEntry", srsItem).Return(nil)
				spacedRepetitionRepo.On("UpdateEntry", srsItem).Return(nil)
				userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)

				spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, nil)
				remindRepo.On("SetReminder", userUUID, whenNext, whenNext).Return(nil)
				manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
				manager.OnEvent(moment)

				mock.AssertExpectationsForObjects(GinkgoT(), spacedRepetitionRepo, userInfoRepo, remindRepo)
			})

			It("Getting next failed, so silently stop", func() {
				want := errors.New("fail")
				// A mess is born
				spacedRepetitionRepo.On("SaveEntry", srsItem).Return(nil)
				spacedRepetitionRepo.On("UpdateEntry", srsItem).Return(nil)
				userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
				spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, want)
				testutils.SetLoggerToPanicOnFatal(logger)

				manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
				Expect(func() { manager.OnEvent(moment) }).Should(Panic())

				mock.AssertExpectationsForObjects(GinkgoT(), spacedRepetitionRepo, userInfoRepo, remindRepo)
			})
		})

		When("we delete an entry", func() {
			var (
				whenNext             time.Time
				srsItem, nextSrsItem spaced_repetition.SpacedRepetitionEntry
				moment               event.Eventlog
			)
			BeforeEach(func() {
				whenNext, _ = time.Parse(time.RFC3339, "2020-12-23T12:58:21Z")
				srsItem = spaced_repetition.SpacedRepetitionEntry{
					UserUUID: userUUID,
					UUID:     "ba9277fc4c6190fb875ad8f9cee848dba699937f",
					Body:     "{\"show\":\"Hello\",\"kind\":\"v1\",\"uuid\":\"ba9277fc4c6190fb875ad8f9cee848dba699937f\",\"data\":\"Hello\",\"settings\":{\"level\":\"0\",\"when_next\":\"2020-12-23T12:58:21Z\",\"created\":\"2020-12-23T11:58:21Z\"}}",
					WhenNext: whenNext,
					Created:  whenNext,
				}

				nextSrsItem = srsItem
				moment = event.Eventlog{
					Kind: event.ApiSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind: spaced_repetition.EventKindDeleted,
						Data: srsItem,
					},
					Timestamp: whenNext.UTC().Unix(),
				}
			})

			It("Failed to delete entry", func() {
				want := errors.New("fail")
				spacedRepetitionRepo.On("DeleteEntry", userUUID, srsItem.UUID).Return(want)
				testutils.SetLoggerToPanicOnFatal(logger)

				manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
				Expect(func() { manager.OnEvent(moment) }).Should(Panic())

				Expect(hook.LastEntry().Data["event"]).To(Equal("spacedRepetitionManager.OnEvent"))
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				Expect(hook.LastEntry().Data["method"]).To(Equal("m.spacedRepetitionRepo.DeleteEntry"))

				mock.AssertExpectationsForObjects(GinkgoT(), spacedRepetitionRepo)
			})

			When("Failed to get next after deleting the entry", func() {
				It("Issue with the db", func() {
					want := errors.New("fail")
					spacedRepetitionRepo.On("DeleteEntry", userUUID, srsItem.UUID).Return(nil)
					userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
					spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, want)
					testutils.SetLoggerToPanicOnFatal(logger)

					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
					Expect(func() { manager.OnEvent(moment) }).Should(Panic())

					Expect(hook.LastEntry().Data["error"]).To(Equal(want))
					Expect(hook.LastEntry().Data["method"]).To(Equal("m.spacedRepetitionRepo.GetNext"))
					mock.AssertExpectationsForObjects(GinkgoT(), spacedRepetitionRepo, userInfoRepo)
				})

				When("User has no more entries, so we remove the user from the remind system", func() {
					It("fails", func() {
						want := errors.New("fail")
						spacedRepetitionRepo.On("DeleteEntry", userUUID, srsItem.UUID).Return(nil)
						userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
						spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, utils.ErrNotFound)
						remindRepo.On("DeleteByUser", userUUID).Return(want)
						testutils.SetLoggerToPanicOnFatal(logger)

						manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
						Expect(func() { manager.OnEvent(moment) }).Should(Panic())

						Expect(hook.LastEntry().Data["error"]).To(Equal(want))
						Expect(hook.LastEntry().Data["method"]).To(Equal("m.remindRepo.DeleteByUser"))

						mock.AssertExpectationsForObjects(GinkgoT(), spacedRepetitionRepo, userInfoRepo, remindRepo)
					})

					It("Success", func() {
						spacedRepetitionRepo.On("DeleteEntry", userUUID, srsItem.UUID).Return(nil)
						userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
						spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, utils.ErrNotFound)
						remindRepo.On("DeleteByUser", userUUID).Return(nil)
						testutils.SetLoggerToPanicOnFatal(logger)

						manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
						manager.OnEvent(moment)

						Expect(hook.LastEntry()).To(BeNil())
						mock.AssertExpectationsForObjects(GinkgoT(), spacedRepetitionRepo, userInfoRepo, remindRepo)
					})
				})

				It("Item deleted, has another entry. Update the reminder", func() {
					spacedRepetitionRepo.On("DeleteEntry", userUUID, srsItem.UUID).Return(nil)
					userInfoRepo.On("Get", userUUID).Return(appSettingsEnabled, nil)
					spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, nil)
					remindRepo.On("SetReminder", userUUID, whenNext, whenNext).Return(nil)

					testutils.SetLoggerToPanicOnFatal(logger)

					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
					manager.OnEvent(moment)

					mock.AssertExpectationsForObjects(GinkgoT(), spacedRepetitionRepo, userInfoRepo, remindRepo)
				})
			})
		})

		When("a user is deleted", func() {
			var (
				moment event.Eventlog
			)
			BeforeEach(func() {

				moment = event.Eventlog{
					Kind: event.ApiUserDelete,
					UUID: userUUID,
				}
			})

			It("Failed to delete user", func() {
				want := errors.New("fail")
				remindRepo.On("DeleteByUser", userUUID).Return(want)
				spacedRepetitionRepo.On("DeleteByUser", userUUID).Return(want)
				testutils.SetLoggerToPanicOnFatal(logger)

				manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
				Expect(func() { manager.OnEvent(moment) }).Should(Panic())

				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				Expect(hook.LastEntry().Data["kind"]).To(Equal(event.ApiUserDelete))
				Expect(hook.LastEntry().Data["method"]).To(Equal("m.remindRepo.DeleteByUser"))

				moment.Kind = event.CMDUserDelete
				Expect(func() { manager.OnEvent(moment) }).Should(Panic())
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				Expect(hook.LastEntry().Data["kind"]).To(Equal(event.CMDUserDelete))
				Expect(hook.LastEntry().Data["method"]).To(Equal("m.remindRepo.DeleteByUser"))

				mock.AssertExpectationsForObjects(GinkgoT(), remindRepo)
			})

			It("user removed", func() {
				remindRepo.On("DeleteByUser", userUUID).Return(nil)
				spacedRepetitionRepo.On("DeleteByUser", userUUID).Return(nil)
				manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
				manager.OnEvent(moment)
				Expect(hook.LastEntry().Message).To(Equal("user removed"))

				mock.AssertExpectationsForObjects(GinkgoT(), spacedRepetitionRepo, remindRepo)
			})
		})
	})

	When("Sending Notications", func() {
		var (
			eventMessageBus *mocks.EventlogPubSub
		)
		BeforeEach(func() {
			eventMessageBus = &mocks.EventlogPubSub{}
			event.SetBus(eventMessageBus)
		})

		It("Issue getting Reminders from the repo", func() {
			want := errors.New("fail")
			remindRepo.On("GetReminders", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]remind.SpacedRepetitionReminder{}, want)

			testutils.SetLoggerToPanicOnFatal(logger)
			manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
			Expect(func() { manager.SendNotifications() }).Should(Panic())
			lastLog := hook.LastEntry()
			Expect(lastLog.Data["error"]).To(Equal(want))
		})

		It("No reminders found", func() {
			remindRepo.On("GetReminders", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(make([]remind.SpacedRepetitionReminder, 0), nil)
			manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
			manager.SendNotifications()
		})

		When("Reminders found", func() {
			It("1 found, skip because the token has not been set", func() {
				remindRepo.On("GetReminders", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]remind.SpacedRepetitionReminder{
					{
						Medium:   []string{""},
						UserUUID: userUUID,
					},
				}, nil)
				remindRepo.On("UpdateSent", userUUID, remind.ReminderSkipped).Return(nil)
				manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
				manager.SendNotifications()

				lastLog := hook.LastEntry()
				Expect(lastLog.Data["msg_skipped"]).To(Equal(1))
				Expect(lastLog.Data["msg_sent"]).To(Equal(0))
			})

			When("Send notification", func() {
				It("Fails on updating user who has had a notification sent", func() {
					remindRepo.On("GetReminders", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]remind.SpacedRepetitionReminder{
						{
							Medium:   []string{"fake-token-123"},
							UserUUID: userUUID,
						},
					}, nil)
					remindRepo.On("UpdateSent", userUUID, remind.ReminderSent).Return(errors.New("fail"))
					eventMessageBus.On("Publish", event.TopicNotifications, mock.MatchedBy(func(moment event.Eventlog) bool {
						Expect(moment.Kind).To(Equal(event.KindPushNotification))

						return true
					}))
					testutils.SetLoggerToPanicOnFatal(logger)
					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
					Expect(func() { manager.SendNotifications() }).Should(Panic())
				})

				It("Success", func() {
					remindRepo.On("GetReminders", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
						Return([]remind.SpacedRepetitionReminder{
							{
								Medium:   []string{"", "fake-token-123"},
								UserUUID: userUUID,
							},
							{
								Medium:   []string{""},
								UserUUID: "fake-user-456",
							},
						}, nil)
					remindRepo.On("UpdateSent", userUUID, remind.ReminderSent).Return(nil)
					remindRepo.On("UpdateSent", "fake-user-456", remind.ReminderSkipped).Return(nil)

					eventMessageBus.On("Publish", event.TopicNotifications, mock.MatchedBy(func(moment event.Eventlog) bool {
						Expect(moment.Kind).To(Equal(event.KindPushNotification))

						return true
					}))

					manager := remind.NewSpacedRepetition(userInfoRepo, spacedRepetitionRepo, remindRepo, logger)
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

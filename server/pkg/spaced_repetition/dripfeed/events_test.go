package dripfeed_test

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Events", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		hook            *test.Hook

		want         error
		service      dripfeed.DripfeedService
		dripfeedRepo *mocks.DripfeedRepository
		listRepo     *mocks.DatastoreAlists
		aclRepo      *mocks.Acl

		userUUID     string
		dripfeedUUID string
		moment       event.Eventlog
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()

		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		eventMessageBus.On("Subscribe", event.TopicMonolog, "dripfeedService", mock.Anything)

		dripfeedRepo = &mocks.DripfeedRepository{}
		aclRepo = &mocks.Acl{}
		listRepo = &mocks.DatastoreAlists{}

		want = errors.New("want")
		userUUID = "fake-user-123"
		dripfeedUUID = "fake-dripfeed-123"
	})

	It("Remove user", func() {
		moment = event.Eventlog{
			Kind: event.ApiUserDelete,
			UUID: userUUID,
		}

		dripfeedRepo.On("DeleteByUser", userUUID).Return(nil)

		service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
		service.OnEvent(moment)
		Expect(hook.LastEntry().Data["event"]).To(Equal(event.UserDeleted))
		Expect(hook.LastEntry().Data["user_uuid"]).To(Equal(userUUID))

		moment.Kind = event.CMDUserDelete
		service.OnEvent(moment)
		Expect(hook.LastEntry().Data["event"]).To(Equal(event.UserDeleted))
		Expect(hook.LastEntry().Data["user_uuid"]).To(Equal(userUUID))
	})

	When("adding a list over time for spaced repetition via api", func() {
		When("Removing a list from dripfeed", func() {
			BeforeEach(func() {
				moment = event.Eventlog{
					Kind: event.ApiDripfeed,
					UUID: userUUID,
					Data: dripfeed.EventDripfeedDelete{
						DripfeedUUID: dripfeedUUID,
						UserUUID:     userUUID,
					},
					Action: event.ActionDeleted,
				}
			})

			It("Issue with the db getting info", func() {
				dripfeedRepo.On("GetInfo", dripfeedUUID).Return(openapi.SpacedRepetitionOvertimeInfo{}, want)
				testutils.SetLoggerToPanicOnFatal(logger)

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				Expect(func() { service.OnEvent(moment) }).Should(Panic())

				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			})

			It("Issue with the db deleting all", func() {
				testutils.SetLoggerToPanicOnFatal(logger)

				dripfeedRepo.On("GetInfo", dripfeedUUID).Return(openapi.SpacedRepetitionOvertimeInfo{
					AlistUuid:    "fake-list-123",
					UserUuid:     userUUID,
					DripfeedUuid: dripfeedUUID,
				}, nil)

				dripfeedRepo.On("DeleteByUUIDAndUserUUID", dripfeedUUID, userUUID).Return(want)
				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)

				Expect(func() { service.OnEvent(moment) }).Should(Panic())

				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				mock.AssertExpectationsForObjects(GinkgoT(), dripfeedRepo)
			})

			It("Success", func() {
				info := openapi.SpacedRepetitionOvertimeInfo{
					AlistUuid:    "fake-list-123",
					UserUuid:     userUUID,
					DripfeedUuid: dripfeedUUID,
				}
				dripfeedRepo.On("GetInfo", dripfeedUUID).Return(info, nil)
				dripfeedRepo.On("DeleteByUUIDAndUserUUID", dripfeedUUID, userUUID).Return(nil)
				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(dripfeed.EventDripfeedRemoved))
					Expect(moment.Data.(openapi.SpacedRepetitionOvertimeInfo)).To(Equal(info))
					return true
				}))

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				service.OnEvent(moment)

			})
		})
	})

	When("a list has been added for dripfeed", func() {
		var alistUUID string
		BeforeEach(func() {
			alistUUID = "fake-list-123"
			dripfeedUUID = dripfeed.UUID(userUUID, alistUUID)

			moment = event.Eventlog{
				UUID:   dripfeedUUID,
				Kind:   event.ApiDripfeed,
				Action: event.ActionCreated,
			}
		})

		When("Fail to add all items due to issue with the repo", func() {
			It("Via V1", func() {
				testutils.SetLoggerToPanicOnFatal(logger)
				aList := alist.Alist{}
				aList.Uuid = alistUUID
				aList.Info.ListType = alist.SimpleList
				aList.Data = make(alist.TypeV1, 0)
				aList.Data = append(aList.Data.(alist.TypeV1), "hello")

				moment.Data = dripfeed.EventDripfeedInputV1{
					Info: dripfeed.EventDripfeedInputBase{
						UserUUID:  userUUID,
						AlistUUID: aList.Uuid,
						Kind:      aList.Info.ListType,
					},
					Data: aList.Data.(alist.TypeV1),
				}

				dripfeedRepo.On("AddAll", dripfeedUUID, userUUID, alistUUID, mock.MatchedBy(func(data []string) bool {
					Expect(len(data)).To(Equal(1))
					// Check for proof of hash
					Expect(strings.Contains(data[0], `"uuid":"a1f2fbfe2c4ad81749cd0380b735295d06f9d0c4"`)).To(BeTrue())
					// Check for dripfeedUUID
					Expect(strings.Contains(data[0], dripfeedUUID)).To(BeTrue())
					return true
				})).Return(want)

				//aList.Data = append(aList.Data.(alist.TypeV1), "hello")
				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)

				Expect(func() { service.OnEvent(moment) }).Should(Panic())
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			})

			It("Via V2", func() {
				testutils.SetLoggerToPanicOnFatal(logger)
				aList := alist.Alist{}
				aList.Uuid = alistUUID
				aList.Info.ListType = alist.FromToList
				aList.Data = make(alist.TypeV2, 0)
				aList.Data = append(aList.Data.(alist.TypeV2), alist.TypeV2Item{
					From: "bil",
					To:   "car",
				})

				moment.Data = dripfeed.EventDripfeedInputV2{
					Info: dripfeed.EventDripfeedInputBase{
						UserUUID:  userUUID,
						AlistUUID: aList.Uuid,
						Kind:      aList.Info.ListType,
					},
					Settings: openapi.SpacedRepetitionOvertimeInputV2AllOfSettings{
						Show: "from",
					},
					Data: aList.Data.(alist.TypeV2),
				}

				dripfeedRepo.On("AddAll", dripfeedUUID, userUUID, alistUUID, mock.MatchedBy(func(data []string) bool {
					Expect(len(data)).To(Equal(1))
					// Check for proof of hash
					Expect(strings.Contains(data[0], `"uuid":"7c3d3151cb6d7f21549c9ecd2976e745eb3ef852"`)).To(BeTrue())
					// Check for dripfeedUUID
					Expect(strings.Contains(data[0], dripfeedUUID)).To(BeTrue())
					return true
				})).Return(want)

				//aList.Data = append(aList.Data.(alist.TypeV1), "hello")
				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)

				Expect(func() { service.OnEvent(moment) }).Should(Panic())
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			})
		})

		When("Success", func() {
			var whenNext time.Time
			BeforeEach(func() {
				aList := alist.Alist{}
				aList.Uuid = alistUUID
				aList.Info.ListType = alist.FromToList
				aList.Data = make(alist.TypeV2, 0)
				aList.Data = append(aList.Data.(alist.TypeV2), alist.TypeV2Item{
					From: "bil",
					To:   "car",
				})

				whenNext = time.Now().UTC()
				moment.Timestamp = whenNext.Unix()
				moment.Data = dripfeed.EventDripfeedInputV2{
					Info: dripfeed.EventDripfeedInputBase{
						UserUUID:  userUUID,
						AlistUUID: aList.Uuid,
						Kind:      aList.Info.ListType,
					},
					Settings: openapi.SpacedRepetitionOvertimeInputV2AllOfSettings{
						Show: "from",
					},
					Data: aList.Data.(alist.TypeV2),
				}

				dripfeedRepo.On("AddAll", dripfeedUUID, userUUID, alistUUID, mock.MatchedBy(func(data []string) bool {
					Expect(len(data)).To(Equal(1))
					// Check for proof of hash
					Expect(strings.Contains(data[0], `"uuid":"7c3d3151cb6d7f21549c9ecd2976e745eb3ef852"`)).To(BeTrue())
					// Check for dripfeedUUID
					Expect(strings.Contains(data[0], dripfeedUUID)).To(BeTrue())
					return true
				})).Return(nil)

			})

			It("Error looking for next item", func() {
				testutils.SetLoggerToPanicOnFatal(logger)
				expectedEvents := []string{}
				verify := func(args mock.Arguments) {
					moment := args[1].(event.Eventlog)
					expectedEvents = append(expectedEvents, moment.Kind)
				}

				eventMessageBus.On("Publish", event.TopicMonolog, mock.Anything).Times(2).Run(verify)
				dripfeedRepo.On("GetNext", dripfeedUUID).Return(dripfeed.RepoItem{}, want)

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				Expect(func() { service.OnEvent(moment) }).Should(Panic())
				Expect(expectedEvents).To(Equal([]string{dripfeed.EventDripfeedAdded}))
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				Expect(hook.LastEntry().Data["method"]).To(Equal("s.checkForNext"))
			})

			It("List was empty, so nothing found", func() {
				expectedEvents := []string{}
				verify := func(args mock.Arguments) {
					moment := args[1].(event.Eventlog)
					expectedEvents = append(expectedEvents, moment.Kind)
				}

				eventMessageBus.On("Publish", event.TopicMonolog, mock.Anything).Times(2).Run(verify)
				dripfeedRepo.On("GetNext", dripfeedUUID).Return(dripfeed.RepoItem{}, utils.ErrNotFound)

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				service.OnEvent(moment)
				Expect(expectedEvents).To(Equal([]string{dripfeed.EventDripfeedAdded, dripfeed.EventDripfeedFinished}))
				mock.AssertExpectationsForObjects(GinkgoT(), eventMessageBus)
			})

			It("V1 entry has been found", func() {
				entry := dripfeed.RepoItem{
					SrsUUID: "ba9277fc4c6190fb875ad8f9cee848dba699937f",
					SrsKind: "v1",
					// TODO get a real example
					SrsBody:      []byte(`{"show":"Hello","kind":"v1","uuid":"ba9277fc4c6190fb875ad8f9cee848dba699937f","data":"Hello","settings":{"level":"0","when_next":"2020-12-27T18:04:59Z","created":"2020-12-27T17:04:59Z","ext_id":"f29a45249551ae992a8edc6526ca7421094c8883"}}`),
					Position:     0,
					DripfeedUUID: dripfeedUUID,
					UserUUID:     userUUID,
					AlistUUID:    alistUUID,
				}

				expectedEvents := []string{}
				verify := func(args mock.Arguments) {
					moment := args[1].(event.Eventlog)
					expectedEvents = append(expectedEvents, moment.Kind)

					if moment.Kind == event.SystemSpacedRepetition {
						Expect(moment.Action).To(Equal(spaced_repetition.EventKindNew))
						srsItem := moment.Data.(spaced_repetition.EventSpacedRepetition).Data
						Expect(srsItem.UUID).To(Equal("ba9277fc4c6190fb875ad8f9cee848dba699937f"))
						// Confirm the time set in the event, is the one used to decide when to reset from
						// Confirm dripfeedUUID is being set as the ext_id
						var expectedSrsItem openapi.SpacedRepetitionV1
						json.Unmarshal([]byte(srsItem.Body), &expectedSrsItem)
						Expect(expectedSrsItem.Settings.ExtId).To(Equal(dripfeedUUID))

						// Confirm the time set in the event, is the one used to decide when to reset from
						Expect(whenNext.Add(spaced_repetition.Threshold0).Format(time.RFC3339)).To(Equal(srsItem.WhenNext.Format(time.RFC3339)))
					}
				}

				eventMessageBus.On("Publish", event.TopicMonolog, mock.Anything).Times(2).Run(verify)
				dripfeedRepo.On("GetNext", dripfeedUUID).Return(entry, nil)

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				service.OnEvent(moment)
				Expect(expectedEvents).To(Equal([]string{dripfeed.EventDripfeedAdded, event.SystemSpacedRepetition}))
			})

			It("V2 entry has been found", func() {
				entry := dripfeed.RepoItem{
					SrsUUID: "75698c0f5a7b904f1799ceb68e2afe67ad987689",
					SrsKind: "v2",
					// TODO get a real example
					SrsBody:      []byte(`{"data":{"from":"March","to":"Mars"},"kind":"v2","settings":{"created":"2020-12-28T11:44:33Z","level":"0","show":"to","when_next":"2020-12-28T12:44:33Z","ext_id":"f29a45249551ae992a8edc6526ca7421094c8883"},"show":"Mars","uuid":"75698c0f5a7b904f1799ceb68e2afe67ad987689"}`),
					Position:     0,
					DripfeedUUID: dripfeedUUID,
					UserUUID:     userUUID,
					AlistUUID:    alistUUID,
				}

				expectedEvents := []string{}
				verify := func(args mock.Arguments) {
					moment := args[1].(event.Eventlog)
					expectedEvents = append(expectedEvents, moment.Kind)

					if moment.Kind == event.SystemSpacedRepetition {
						Expect(moment.Action).To(Equal(spaced_repetition.EventKindNew))
						srsItem := moment.Data.(spaced_repetition.EventSpacedRepetition).Data
						Expect(srsItem.UUID).To(Equal(entry.SrsUUID))

						// Confirm dripfeedUUID is being set as the ext_id
						var expectedSrsItem openapi.SpacedRepetitionV2
						json.Unmarshal([]byte(srsItem.Body), &expectedSrsItem)
						Expect(expectedSrsItem.Settings.ExtId).To(Equal(dripfeedUUID))

						// Confirm the time set in the event, is the one used to decide when to reset from
						Expect(whenNext.Add(spaced_repetition.Threshold0).Format(time.RFC3339)).To(Equal(srsItem.WhenNext.Format(time.RFC3339)))
					}
				}

				eventMessageBus.On("Publish", event.TopicMonolog, mock.Anything).Times(2).Run(verify)
				dripfeedRepo.On("GetNext", dripfeedUUID).Return(entry, nil)

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				service.OnEvent(moment)
				Expect(expectedEvents).To(Equal([]string{dripfeed.EventDripfeedAdded, event.SystemSpacedRepetition}))
			})
		})
	})

	When("Spaced Repetition triggers event", func() {
		var (
			whenNext time.Time
			srsItem  spaced_repetition.SpacedRepetitionEntry
			moment   event.Eventlog
		)

		Context("New entry", func() {
			BeforeEach(func() {
				whenNext, _ = time.Parse(time.RFC3339, "2020-12-23T12:58:21Z")
				srsItem = spaced_repetition.SpacedRepetitionEntry{
					UserUUID: userUUID,
					UUID:     "ba9277fc4c6190fb875ad8f9cee848dba699937f",
					Body:     "{\"show\":\"Hello\",\"kind\":\"v1\",\"uuid\":\"ba9277fc4c6190fb875ad8f9cee848dba699937f\",\"data\":\"Hello\",\"settings\":{\"level\":\"0\",\"when_next\":\"2020-12-23T12:58:21Z\",\"created\":\"2020-12-23T11:58:21Z\"}}",
					WhenNext: whenNext,
					Created:  whenNext,
				}

				moment = event.Eventlog{
					Kind: event.ApiSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind: spaced_repetition.EventKindNew,
						Data: srsItem,
					},
					Timestamp: whenNext.UTC().Unix(),
				}
			})

			It("Issue with repo when deleting entry", func() {
				testutils.SetLoggerToPanicOnFatal(logger)

				dripfeedRepo.On("DeleteAllByUserUUIDAndSpacedRepetitionUUID", userUUID, srsItem.UUID).Return(want)

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)

				Expect(func() { service.OnEvent(moment) }).Should(Panic())
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				mock.AssertExpectationsForObjects(GinkgoT(), dripfeedRepo)
			})

			It("Success", func() {
				dripfeedRepo.On("DeleteAllByUserUUIDAndSpacedRepetitionUUID", userUUID, srsItem.UUID).Return(nil).Times(1)

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)

				service.OnEvent(moment)
			})
		})

		Context("Entry Viewed", func() {
			var (
				entry openapi.SpacedRepetitionV1
			)

			BeforeEach(func() {
				created, _ := time.Parse(time.RFC3339, "2020-12-23T11:58:21Z")
				whenNext, _ = time.Parse(time.RFC3339, "2020-12-23T12:58:21Z")
				entry = openapi.SpacedRepetitionV1{
					Show: "Hello",
					Kind: alist.SimpleList,
					Uuid: "ba9277fc4c6190fb875ad8f9cee848dba699937f",
					Data: "Hello",
					Settings: openapi.SpacedRepetitionBaseSettings{
						Level:    "0",
						Created:  created,
						WhenNext: whenNext,
					},
				}
			})

			It("Not increment", func() {
				entryB, _ := json.Marshal(entry)
				srsItem = spaced_repetition.SpacedRepetitionEntry{
					UserUUID: userUUID,
					UUID:     entry.Uuid,
					Body:     string(entryB),
					WhenNext: whenNext,
					Created:  whenNext,
				}

				moment = event.Eventlog{
					Kind: event.ApiSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind:   spaced_repetition.EventKindViewed,
						Data:   srsItem,
						Action: spaced_repetition.ActionDecrement,
					},
					Timestamp: whenNext.UTC().Unix(),
				}

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				service.OnEvent(moment)
			})

			It("Doesnt have an ext_id linked to it", func() {
				entryB, _ := json.Marshal(entry)
				srsItem = spaced_repetition.SpacedRepetitionEntry{
					UserUUID: userUUID,
					UUID:     entry.Uuid,
					Body:     string(entryB),
					WhenNext: whenNext,
					Created:  whenNext,
				}

				moment = event.Eventlog{
					Kind: event.ApiSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind:   spaced_repetition.EventKindViewed,
						Data:   srsItem,
						Action: spaced_repetition.ActionIncrement,
					},
					Timestamp: whenNext.UTC().Unix(),
				}

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				service.OnEvent(moment)
			})

			It("Not level1", func() {
				entry.Settings.ExtId = "fake-dripfeed-123"
				entryB, _ := json.Marshal(entry)
				srsItem = spaced_repetition.SpacedRepetitionEntry{
					UserUUID: userUUID,
					UUID:     entry.Uuid,
					Body:     string(entryB),
					WhenNext: whenNext,
					Created:  whenNext,
				}

				moment = event.Eventlog{
					Kind: event.ApiSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind:   spaced_repetition.EventKindViewed,
						Data:   srsItem,
						Action: spaced_repetition.ActionIncrement,
					},
					Timestamp: whenNext.UTC().Unix(),
				}

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				service.OnEvent(moment)
			})

			Context("ExtID / DripfeedUUID is not found", func() {
				BeforeEach(func() {
					dripfeedUUID = "fake-dripfeed-123"
					entry.Settings.ExtId = dripfeedUUID
					entry.Settings.Level = "1"

					entryB, _ := json.Marshal(entry)
					srsItem = spaced_repetition.SpacedRepetitionEntry{
						UserUUID: userUUID,
						UUID:     entry.Uuid,
						Body:     string(entryB),
						WhenNext: whenNext,
						Created:  whenNext,
					}

					moment = event.Eventlog{
						Kind: event.ApiSpacedRepetition,
						Data: spaced_repetition.EventSpacedRepetition{
							Kind:   spaced_repetition.EventKindViewed,
							Data:   srsItem,
							Action: spaced_repetition.ActionIncrement,
						},
						Timestamp: whenNext.UTC().Unix(),
					}
				})
				It("Issue talking to repo", func() {
					dripfeedRepo.On("GetInfo", dripfeedUUID).Return(openapi.SpacedRepetitionOvertimeInfo{}, want).Once()
					testutils.SetLoggerToPanicOnFatal(logger)

					service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
					Expect(func() { service.OnEvent(moment) }).Should(Panic())

					Expect(hook.LastEntry().Data["error"]).To(Equal(want))
				})

				It("Success", func() {
					dripfeedRepo.On("GetInfo", dripfeedUUID).Return(openapi.SpacedRepetitionOvertimeInfo{}, utils.ErrNotFound).Once()
					service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
					service.OnEvent(moment)
				})
			})

			It("Look up for next", func() {
				// TODO or remove
			})
		})
	})

	When("The system triggers a spaced repetition event", func() {
		var (
			whenNext time.Time
			srsItem  spaced_repetition.SpacedRepetitionEntry
			moment   event.Eventlog
			entry    openapi.SpacedRepetitionV1
		)

		BeforeEach(func() {
			created, _ := time.Parse(time.RFC3339, "2020-12-23T11:58:21Z")
			whenNext, _ = time.Parse(time.RFC3339, "2020-12-23T12:58:21Z")
			entry = openapi.SpacedRepetitionV1{
				Show: "Hello",
				Kind: alist.SimpleList,
				Uuid: "ba9277fc4c6190fb875ad8f9cee848dba699937f",
				Data: "Hello",
				Settings: openapi.SpacedRepetitionBaseSettings{
					Level:    "0",
					Created:  created,
					WhenNext: whenNext,
				},
			}
		})

		It("Skip if not already in the system", func() {
			moment = event.Eventlog{
				Kind:   event.SystemSpacedRepetition,
				Action: event.ActionCreated,
			}
			service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
			service.OnEvent(moment)
		})

		It("Issue when trying to remove entry via the repo", func() {
			entryB, _ := json.Marshal(entry)
			srsItem = spaced_repetition.SpacedRepetitionEntry{
				UserUUID: userUUID,
				UUID:     entry.Uuid,
				Body:     string(entryB),
				WhenNext: whenNext,
				Created:  whenNext,
			}

			moment = event.Eventlog{
				Kind: event.SystemSpacedRepetition,
				Data: spaced_repetition.EventSpacedRepetition{
					Data: srsItem,
				},
				Action:    spaced_repetition.EventKindAlreadyInSystem,
				Timestamp: whenNext.UTC().Unix(),
			}
			dripfeedRepo.On("DeleteAllByUserUUIDAndSpacedRepetitionUUID", userUUID, srsItem.UUID).Return(want).Times(1)

			testutils.SetLoggerToPanicOnFatal(logger)
			service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
			Expect(func() { service.OnEvent(moment) }).Should(Panic())
			Expect(hook.LastEntry().Data["error"]).To(Equal(want))
		})

		It("SRS Entry doesnt have ext_id", func() {
			entry.Settings.ExtId = ""
			entryB, _ := json.Marshal(entry)
			srsItem = spaced_repetition.SpacedRepetitionEntry{
				UserUUID: userUUID,
				UUID:     entry.Uuid,
				Body:     string(entryB),
				WhenNext: whenNext,
				Created:  whenNext,
			}

			moment = event.Eventlog{
				Kind: event.SystemSpacedRepetition,
				Data: spaced_repetition.EventSpacedRepetition{
					Data: srsItem,
				},
				Action:    spaced_repetition.EventKindAlreadyInSystem,
				Timestamp: whenNext.UTC().Unix(),
			}
			dripfeedRepo.On("DeleteAllByUserUUIDAndSpacedRepetitionUUID", userUUID, srsItem.UUID).Return(nil).Times(1)

			service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
			service.OnEvent(moment)
		})

		Context("ExtID / DripfeedUUID is not found", func() {

			BeforeEach(func() {
				dripfeedUUID = "fake-dripfeed-123"
				entry.Settings.ExtId = dripfeedUUID
				entry.Settings.Level = "1"

				entryB, _ := json.Marshal(entry)
				srsItem = spaced_repetition.SpacedRepetitionEntry{
					UserUUID: userUUID,
					UUID:     entry.Uuid,
					Body:     string(entryB),
					WhenNext: whenNext,
					Created:  whenNext,
				}

				moment = event.Eventlog{
					Kind: event.SystemSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Data: srsItem,
					},
					Action:    spaced_repetition.EventKindAlreadyInSystem,
					Timestamp: whenNext.UTC().Unix(),
				}
				dripfeedRepo.On("DeleteAllByUserUUIDAndSpacedRepetitionUUID", userUUID, srsItem.UUID).Return(nil).Times(1)
			})

			It("Issue talking to repo", func() {
				dripfeedRepo.On("GetInfo", dripfeedUUID).Return(openapi.SpacedRepetitionOvertimeInfo{}, want).Once()
				testutils.SetLoggerToPanicOnFatal(logger)

				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				Expect(func() { service.OnEvent(moment) }).Should(Panic())

				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			})

			It("Success", func() {
				dripfeedRepo.On("GetInfo", dripfeedUUID).Return(openapi.SpacedRepetitionOvertimeInfo{}, utils.ErrNotFound).Once()
				service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
				service.OnEvent(moment)
			})
		})
	})

	When("Adding overtime is finished", func() {
		BeforeEach(func() {
			moment = event.Eventlog{
				Kind: dripfeed.EventDripfeedFinished,
				Data: openapi.SpacedRepetitionOvertimeInfo{
					DripfeedUuid: dripfeedUUID,
					UserUuid:     userUUID,
					AlistUuid:    "fake-list-123",
				},

				UUID: dripfeedUUID,
			}
		})

		It("failed to remove info, issue talking to repo", func() {
			testutils.SetLoggerToPanicOnFatal(logger)
			dripfeedRepo.On("DeleteByUUIDAndUserUUID", dripfeedUUID, userUUID).Return(want).Once()
			service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)

			Expect(func() { service.OnEvent(moment) }).Should(Panic())
			Expect(hook.LastEntry().Data["error"]).To(Equal(want))
		})

		It("Info removed", func() {
			dripfeedRepo.On("DeleteByUUIDAndUserUUID", dripfeedUUID, userUUID).Return(nil).Once()
			service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
			service.OnEvent(moment)
		})
	})
})

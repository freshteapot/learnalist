package dripfeed_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
					fmt.Println(data)
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

			BeforeEach(func() {
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
					fmt.Println(data)
					Expect(len(data)).To(Equal(1))
					// Check for proof of hash
					Expect(strings.Contains(data[0], `"uuid":"7c3d3151cb6d7f21549c9ecd2976e745eb3ef852"`)).To(BeTrue())
					// Check for dripfeedUUID
					Expect(strings.Contains(data[0], dripfeedUUID)).To(BeTrue())
					return true
				})).Return(nil)

				//aList.Data = append(aList.Data.(alist.TypeV1), "hello")

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
		})
	})

	It("REMOVE", func() {
		Expect("1").To(Equal("1"))
	})

	It("Making sure things work", func() {
		type eventDripfeedInput struct {
			UserUUID string      `json:"user_uuid"`
			Kind     string      `json:"kind"` // This is the list_type, at some point I will drop list_type :P
			Data     interface{} `json:"data"` // TODO I think with openapi-generator we might be able to move to something else.
		}
		raw := `{"kind":"api.dripfeed","data":{"user_uuid":"7197b389-cfe6-4fa8-9aea-98d49b305039","kind":"v1","data":["monday","tuesday","wednesday","thursday","friday","saturday","sunday"]},"timestamp":1610273658,"action":"created"}`

		var entry event.Eventlog
		json.Unmarshal([]byte(raw), &entry)
		b, _ := json.Marshal(entry.Data)

		var moment dripfeed.EventDripfeedInputV1
		json.Unmarshal(b, &moment)
		fmt.Println(moment.Data)
	})

	It("Check if new has dripfeed", func() {
		raw := `{"kind":"api.spacedrepetition","data":{"kind":"new","data":{"uuid":"bfe3cc8ad82c1e8282b53df0a7a78685042d9f5b","body":"{\"show\":\"monday\",\"kind\":\"v1\",\"uuid\":\"bfe3cc8ad82c1e8282b53df0a7a78685042d9f5b\",\"data\":\"monday\",\"settings\":{\"level\":\"0\",\"when_next\":\"2021-01-10T15:37:28Z\",\"created\":\"2021-01-10T14:37:28Z\",\"ext_id\":\"b17ef2deb2d1836dfe534de67e710e23c5b67e88\"}}","user_uuid":"4eccc98d-90ea-42ba-84d4-d0688b64d24e","when_next":"2021-01-10T15:37:28Z","created":"2021-01-10T14:37:28Z"}},"timestamp":1610289448}`
		var entry event.Eventlog
		json.Unmarshal([]byte(raw), &entry)

		b, _ := json.Marshal(entry.Data)
		var moment spaced_repetition.EventSpacedRepetition
		json.Unmarshal(b, &moment)

		srsItem := moment.Data

		var info dripfeed.SpacedRepetitionSettingsExtID

		json.Unmarshal([]byte(srsItem.Body), &info)
		fmt.Println(info.Settings.ExtID)
	})
})

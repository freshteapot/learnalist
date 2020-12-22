package remind_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Spaced Repetition Manager", func() {
	var (
		logger *logrus.Logger
		//mockSql      sqlmock.Sqlmock

		spacedRepetitionRepo           *mocks.SpacedRepetitionRepository
		remindRepo                     *mocks.RemindSpacedRepetitionRepository
		deviceInfo                     openapi.MobileDeviceInfo
		userUUID, appIdentifier, token string
	)

	BeforeEach(func() {
		logger, _ = test.NewNullLogger()
		spacedRepetitionRepo = &mocks.SpacedRepetitionRepository{}
		remindRepo = &mocks.RemindSpacedRepetitionRepository{}
		userUUID = "fake-user-123"
		appIdentifier = apps.RemindV1
		token = "fake-token-123"
		deviceInfo = openapi.MobileDeviceInfo{
			UserUuid:      userUUID,
			AppIdentifier: appIdentifier,
			Token:         token,
		}
		fmt.Println(deviceInfo)
		Expect("A").To(Equal("A"))
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
				spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, nil)
				remindRepo.On("SetReminder", userUUID, whenNext, whenNext).Return(nil)
				manager := remind.NewSpacedRepetition(spacedRepetitionRepo, remindRepo, logger)
				manager.OnEvent(moment)

				mock.AssertExpectationsForObjects(GinkgoT())
			})

			It("Getting next failed, so silently stop", func() {
				want := errors.New("fail")
				// A mess is born
				spacedRepetitionRepo.On("SaveEntry", srsItem).Return(nil)
				spacedRepetitionRepo.On("UpdateEntry", srsItem).Return(nil)
				spacedRepetitionRepo.On("GetNext", userUUID).Return(nextSrsItem, want)
				manager := remind.NewSpacedRepetition(spacedRepetitionRepo, remindRepo, logger)
				manager.OnEvent(moment)

				mock.AssertExpectationsForObjects(GinkgoT())
			})
		})

	})

})

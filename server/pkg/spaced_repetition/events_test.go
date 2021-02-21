package spaced_repetition_test

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
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

		want                 error
		service              spaced_repetition.SpacedRepetitionService
		spacedRepetitionRepo *mocks.SpacedRepetitionRepository

		userUUID string
		moment   event.Eventlog

		// NewService(repo SpacedRepetitionRepository, logContext logrus.FieldLogger) SpacedRepetitionService
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()

		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		eventMessageBus.On("Subscribe", event.TopicMonolog, "spacedRepetitionService", mock.Anything)

		spacedRepetitionRepo = &mocks.SpacedRepetitionRepository{}

		want = errors.New("want")
		userUUID = "fake-user-123"
	})

	It("Remove user", func() {
		moment = event.Eventlog{
			Kind: event.ApiUserDelete,
			UUID: userUUID,
		}

		spacedRepetitionRepo.On("DeleteByUser", userUUID).Return(nil)

		service = spaced_repetition.NewService(spacedRepetitionRepo, logger)
		service.OnEvent(moment)
		Expect(hook.LastEntry().Data["event"]).To(Equal(event.UserDeleted))
		Expect(hook.LastEntry().Data["user_uuid"]).To(Equal(userUUID))

		moment.Kind = event.CMDUserDelete
		service.OnEvent(moment)
		Expect(hook.LastEntry().Data["event"]).To(Equal(event.UserDeleted))
		Expect(hook.LastEntry().Data["user_uuid"]).To(Equal(userUUID))
	})

	When("SystemSpacedRepetition event", func() {
		var (
			moment event.Eventlog
		)

		It("Not EventKindNew", func() {
			moment = event.Eventlog{
				Kind: event.SystemSpacedRepetition,
				Data: spaced_repetition.EventSpacedRepetition{
					Kind: event.ActionCreated,
				},
			}
			service = spaced_repetition.NewService(spacedRepetitionRepo, logger)
			service.OnEvent(moment)
		})

		When("Saving", func() {
			var (
				whenNext time.Time
				srsItem  spaced_repetition.SpacedRepetitionEntry
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
						Kind: spaced_repetition.EventKindNew,
						Data: srsItem,
					},
				}

			})

			It("issue saving", func() {
				spacedRepetitionRepo.On("SaveEntry", mock.Anything).Return(want)

				testutils.SetLoggerToPanicOnFatal(logger)
				service = spaced_repetition.NewService(spacedRepetitionRepo, logger)
				Expect(func() { service.OnEvent(moment) }).Should(Panic())
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			})

			It("Already in the system", func() {
				spacedRepetitionRepo.On("SaveEntry", mock.Anything).Return(spaced_repetition.ErrSpacedRepetitionEntryExists)

				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.SystemSpacedRepetition))
					Expect(moment.Data.(spaced_repetition.EventSpacedRepetition).Kind).To(Equal(spaced_repetition.EventKindAlreadyInSystem))
					Expect(moment.Data.(spaced_repetition.EventSpacedRepetition).Data.UUID).To(Equal(srsItem.UUID))
					return true
				}))

				service = spaced_repetition.NewService(spacedRepetitionRepo, logger)
				service.OnEvent(moment)
			})

			It("Saved", func() {
				spacedRepetitionRepo.On("SaveEntry", mock.Anything).Return(nil)
				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.ApiSpacedRepetition))
					Expect(moment.Data.(spaced_repetition.EventSpacedRepetition).Kind).To(Equal(spaced_repetition.EventKindNew))
					Expect(moment.Data.(spaced_repetition.EventSpacedRepetition).Data.UUID).To(Equal(srsItem.UUID))
					return true
				}))

				service = spaced_repetition.NewService(spacedRepetitionRepo, logger)
				service.OnEvent(moment)
			})
		})

	})
})

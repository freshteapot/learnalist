package challenge_test

import (
	"encoding/json"
	"fmt"

	"firebase.google.com/go/v4/messaging"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Processing push notifications", func() {
	var (
		logger                              *logrus.Logger
		hook                                *test.Hook
		challengePushNotificationRepository *mocks.ChallengePushNotificationRepository
		challengeRepository                 *mocks.ChallengeRepository
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		fmt.Println("Hook", hook)
	})

	It("Quick check to confirm logic for push notifications when success", func() {

		rawMoment := `{"kind":"challenge.newrecord","data":{"kind":"plank","uuid":"07c59b8e-ff54-4a32-8a00-caeebdee523d","user_uuid":"e1848e0b-c939-435e-8090-2f28eb9a2308","data":{"beginningTime":1605972505559,"currentTime":1605972505559,"intervalTime":0,"intervalTimerNow":0,"laps":0,"showIntervals":false,"timerNow":1823,"uuid":"af61b8d8c2422ede274143cd7b5e60916fd20dbf"}}}`
		var entry event.Eventlog
		json.Unmarshal([]byte(rawMoment), &entry)

		challengeRepository = &mocks.ChallengeRepository{}
		challengePushNotificationRepository = &mocks.ChallengePushNotificationRepository{}
		eventMessageBus := &mocks.EventlogPubSub{}

		eventMessageBus.On("Subscribe", event.TopicMonolog, "challenge", mock.Anything)

		eventMessageBus.On("Publish", event.TopicNotifications, mock.MatchedBy(func(moment event.Eventlog) bool {
			Expect(moment.Kind).To(Equal(event.KindPushNotification))

			var msg *messaging.Message
			b, _ := json.Marshal(moment.Data)
			json.Unmarshal(b, &msg)

			Expect(msg.Notification.Title).To(Equal("Challenge update"))
			Expect(msg.Notification.Body).To(Equal("Chris added a plank to A test challenge"))
			return true
		}))
		event.SetBus(eventMessageBus)

		// TODO maybe verify not suppoorted kind
		challengeRepository.On("Get", "07c59b8e-ff54-4a32-8a00-caeebdee523d").Return(challenge.ChallengeInfo{
			UUID:        "07c59b8e-ff54-4a32-8a00-caeebdee523d",
			Kind:        challenge.KindPlankGroup,
			Description: "A test challenge",
		}, nil)

		challengePushNotificationRepository.On("GetUserDisplayName", "e1848e0b-c939-435e-8090-2f28eb9a2308").Return("Chris")
		challengePushNotificationRepository.On("GetUsersInfo", "07c59b8e-ff54-4a32-8a00-caeebdee523d", challenge.PlankGroupMobileApps).Return([]challenge.ChallengeNotificationUserInfo{
			{
				UserUUID: "fake-user-123",
				Token:    "fake-token-123",
			},
		}, nil)

		service := challenge.NewService(challengeRepository, challengePushNotificationRepository, nil, logger)
		service.OnEvent(entry)
	})
})

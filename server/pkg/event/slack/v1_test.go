package event_test

import (
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	eventReader "github.com/freshteapot/learnalist-api/server/pkg/event/slack"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/slack-go/slack"
)

var _ = Describe("Testing Events to Slack", func() {

	var (
		logger  *logrus.Logger
		webhook string
	)

	BeforeEach(func() {
		webhook = ""
		logger, _ = test.NewNullLogger()
	})

	It("Event unwrapping", func() {
		challengeUUID := "fake-challenge-123"
		userUUID := "fake-user-123"

		tests := []struct {
			entry event.Eventlog
			post  eventReader.PostWebhook
		}{
			{
				entry: event.Eventlog{
					Kind: challenge.EventChallengeCreated,
					Data: event.EventKV{
						UUID: challengeUUID,
						Data: challenge.ChallengeInfo{
							UUID:      challengeUUID,
							CreatedBy: userUUID,
						},
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "user:fake-user-123 created challenge:fake-challenge-123"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: challenge.EventChallengeJoined,
					Data: event.EventKV{
						UUID: challengeUUID,
						Data: challenge.ChallengeJoined{
							UUID:     challengeUUID,
							UserUUID: userUUID,
						},
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "user:fake-user-123 joined challenge:fake-challenge-123"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: challenge.EventChallengeLeft,
					Data: event.EventKV{
						UUID: challengeUUID,
						Data: challenge.ChallengeLeft{
							UUID:     challengeUUID,
							UserUUID: userUUID,
						},
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "user:fake-user-123 left challenge:fake-challenge-123"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: mobile.EventMobileDeviceRegistered,
					Data: event.EventKV{
						UUID: userUUID,
						Data: openapi.MobileDeviceInfo{
							UserUuid:      userUUID,
							Token:         "fake",
							AppIdentifier: "remind:v1",
						},
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "user:fake-user-123 registered mobile token"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
		}

		for _, test := range tests {
			reader := eventReader.NewSlackV1Events(test.post, webhook, logger.WithField("context", "slack-events"))
			reader.Read(test.entry)
		}

	})
})

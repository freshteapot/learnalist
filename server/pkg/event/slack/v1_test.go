package event_test

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	eventReader "github.com/freshteapot/learnalist-api/server/pkg/event/slack"
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
		alistUUID := "fake-list-123"

		tests := []struct {
			entry event.Eventlog
			post  eventReader.PostWebhook
		}{
			{
				entry: event.Eventlog{
					Kind: "TODO",
					Data: "I am fake",
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "TODO"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.ApiUserRegister,
					Data: event.EventUser{
						UUID: userUUID,
						Kind: event.KindUserRegisterIDPGoogle,
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "api.user.register: user:fake-user-123 registered via idp:google"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.ApiUserLogin,
					Data: event.EventUser{
						UUID: userUUID,
						Kind: event.KindUserRegisterIDPGoogle,
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "api.user.login: user:fake-user-123 logged in via idp:google"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.ApiUserLogout,
					Data: event.EventUser{
						UUID: userUUID,
						Kind: event.KindUserLogoutSession,
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "api.user.logout: user:fake-user-123 logged out via api, clearing current session"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.ApiUserLogout,
					Data: event.EventUser{
						UUID: userUUID,
						Kind: event.KindUserLogoutSessions,
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "api.user.logout: user:fake-user-123 logged out via api, clearing all sessions"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.BrowserUserLogout,
					Data: event.EventUser{
						UUID: userUUID,
						Kind: event.KindUserLogoutSession,
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "browser.user.logout: user:fake-user-123 logged out via browser, clearing current session"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.BrowserUserLogout,
					Data: event.EventUser{
						UUID: userUUID,
						Kind: event.KindUserLogoutSessions,
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "browser.user.logout: user:fake-user-123 logged out via browser, clearing all sessions"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					UUID: userUUID,
					Kind: event.ApiUserDelete,
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "api.user.delete: user:fake-user-123 should be deleted"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					UUID: userUUID,
					Kind: event.CMDUserDelete,
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "cmd.user.delete: user:fake-user-123 should be deleted"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.ApiListSaved,
					Data: event.EventList{
						UUID:     alistUUID,
						UserUUID: userUUID,
						Action:   "created",
						Data: alist.Alist{
							Info: alist.AlistInfo{
								ListType:   alist.SimpleList,
								SharedWith: keys.SharedWithPublic,
							},
							Data: []string{},
						},
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "list:fake-list-123 (public) created by user:fake-user-123"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.ApiListDelete,
					Data: event.EventList{
						UUID:     alistUUID,
						UserUUID: userUUID,
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "list:fake-list-123 deleted by user:fake-user-123"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
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
					UUID: userUUID,
					Kind: event.MobileDeviceRegistered,
					Data: openapi.MobileDeviceInfo{
						UserUuid:      userUUID,
						Token:         "fake",
						AppIdentifier: "remind:v1",
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `user:fake-user-123 registered mobile token for app:"remind:v1"`
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

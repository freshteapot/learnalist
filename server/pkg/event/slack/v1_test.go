package event_test

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	eventReader "github.com/freshteapot/learnalist-api/server/pkg/event/slack"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
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
		plankUUID := "fake-plank-123"
		dripfeedUUID := "fake-dripfeed-123"

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
					Kind: event.ApiUserDelete,
					UUID: userUUID,
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "api.user.delete: user:fake-user-123 should be deleted"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			// start:event.ApiPlank
			{
				entry: event.Eventlog{
					Kind: event.ApiPlank,
					Data: event.EventPlank{
						Action:   event.ActionNew,
						UserUUID: userUUID,
						Data: openapi.Plank{
							Uuid: plankUUID,
						},
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "user:fake-user-123 added a plank:fake-plank-123"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.ApiPlank,
					Data: event.EventPlank{
						Action:   event.ActionDeleted,
						UserUUID: userUUID,
						Data: openapi.Plank{
							Uuid: plankUUID,
						},
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "user:fake-user-123 deleted plank:fake-plank-123"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.ApiPlank,
					Data: event.EventPlank{
						Action: "not-supported",
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := "not-supported action not supported api.plank"
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			// finish:event.ApiPlank
			{
				entry: event.Eventlog{
					Kind: event.CMDUserDelete,
					UUID: userUUID,
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
						AppIdentifier: apps.RemindV1,
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `user:fake-user-123 registered mobile token for app:"remind_v1"`
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			// start:spaced-repetition.overtime
			{
				entry: event.Eventlog{
					UUID: dripfeedUUID,
					Kind: event.ApiSpacedRepetitionOvertime,
					Data: dripfeed.EventDripfeedInputV1{
						Info: dripfeed.EventDripfeedInputBase{
							AlistUUID: alistUUID,
							UserUUID:  userUUID,
							Kind:      alist.SimpleList,
						},
						Data: make(alist.TypeV1, 0),
					},

					Action: event.ActionCreated,
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `spaced repetition overtime created uuid:fake-dripfeed-123, user:fake-user-123, list:fake-list-123`
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					UUID: dripfeedUUID,
					Kind: event.ApiSpacedRepetitionOvertime,
					Data: openapi.SpacedRepetitionOvertimeInfo{
						DripfeedUuid: dripfeedUUID,
						AlistUuid:    alistUUID,
						UserUuid:     userUUID,
					},

					Action: event.ActionDeleted,
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `spaced repetition overtime deleted uuid:fake-dripfeed-123, user:fake-user-123, list:fake-list-123`
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind:   event.ApiSpacedRepetitionOvertime,
					Action: "not-supported",
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `not-supported action not supported for api.spacedrepetition.overtime`
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.SystemSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind: spaced_repetition.EventKindNew,
						Data: spaced_repetition.SpacedRepetitionEntry{
							UserUUID: userUUID,
						},
					},
					UUID: "fake-srs-item-123",
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `spaced repetition overtime system added entry:fake-srs-item-123 for user:fake-user-123`
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.SystemSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind: spaced_repetition.EventKindAlreadyInSystem,
						Data: spaced_repetition.SpacedRepetitionEntry{
							UserUUID: userUUID,
						},
					},
					UUID: "fake-srs-item-123",
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `spaced repetition overtime system added entry:fake-srs-item-123 for user:fake-user-123 that already exists`
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: event.SystemSpacedRepetition,
					Data: spaced_repetition.EventSpacedRepetition{
						Kind: "not-supported",
					},
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `not-supported kind not supported for system.spacedRepetition`

					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: dripfeed.EventDripfeedAdded,
					Data: openapi.SpacedRepetitionOvertimeInfo{
						DripfeedUuid: dripfeedUUID,
						UserUuid:     userUUID,
						AlistUuid:    alistUUID,
					},
					UUID: dripfeedUUID,
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `spaced repetition overtime activated for user:fake-user-123 from list:fake-list-123`
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: dripfeed.EventDripfeedRemoved,
					Data: openapi.SpacedRepetitionOvertimeInfo{
						DripfeedUuid: dripfeedUUID,
						UserUuid:     userUUID,
						AlistUuid:    alistUUID,
					},
					UUID: dripfeedUUID,
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `spaced repetition overtime stopped for user:fake-user-123 from list:fake-list-123`
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			{
				entry: event.Eventlog{
					Kind: dripfeed.EventDripfeedFinished,
					Data: openapi.SpacedRepetitionOvertimeInfo{
						DripfeedUuid: dripfeedUUID,
						UserUuid:     userUUID,
						AlistUuid:    alistUUID,
					},
					UUID: dripfeedUUID,
				},
				post: func(url string, msg *slack.WebhookMessage) error {
					expect := `spaced repetition overtime finished for user:fake-user-123 from list:fake-list-123`
					Expect(msg.Text).To(Equal(expect))
					return nil
				},
			},
			// finish:spaced-repetition.overtime
		}

		for _, test := range tests {
			reader := eventReader.NewSlackV1Events(test.post, webhook, logger.WithField("context", "slack-events"))
			reader.Read(test.entry)
		}

	})
})

package hugo_test

import (
	"os"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/staticsite/hugo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

// This is all a bit ugly, as the writers are not fully, extracted
// but the logic is still worth mapping thru, only works because external = true
// otherwise it would break a lot
var _ = Describe("Testing Events", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		hook            *test.Hook

		moment     event.Eventlog
		hugoHelper hugo.HugoSiteBuilder
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		eventMessageBus.On("Subscribe", event.TopicMonolog, "hugoHelper", mock.Anything)

		os.MkdirAll("./testdata/content", 0775)
		os.MkdirAll("./testdata/content/alist", 0775)
		os.MkdirAll("./testdata/content/alistsbyuser", 0775)
		os.MkdirAll("./testdata/content/challenge", 0775)
		os.MkdirAll("./testdata/data", 0775)
		os.MkdirAll("./testdata/data/alist", 0775)
		os.MkdirAll("./testdata/data/alistsbyuser", 0775)
		os.MkdirAll("./testdata/data/challenge", 0775)
		os.MkdirAll("./testdata/public", 0775)
		os.MkdirAll("./testdata/public/alist", 0775)
		os.MkdirAll("./testdata/public/alistsbyuser", 0775)
		os.MkdirAll("./testdata/public/challenge", 0775)

		hugoHelper = hugo.NewHugoHelper("./testdata", "fake", logger)
	})

	It("Not supported", func() {
		moment.Kind = "fake"
		hugoHelper.OnEvent(moment)
		Expect(hook.LastEntry().Data["kind"]).To(Equal("fake"))
		Expect(hook.LastEntry().Message).To(Equal("not supported"))
	})

	When("A list has been created / saved or deleted", func() {
		var (
			alistUUID string
		)

		BeforeEach(func() {

			alistUUID = "fake-list-123"

			moment = event.Eventlog{
				Kind: event.ChangesetAlistList,
				UUID: alistUUID,
				Data: alist.Alist{
					Info: alist.AlistInfo{
						ListType:   alist.SimpleList,
						SharedWith: keys.SharedWithPublic,
					},
					Data: []string{},
				},
				Action: event.ActionCreated,
			}
		})

		It("Deleted", func() {
			moment.Action = event.ActionDeleted
			hugoHelper.OnEvent(moment)
			Expect(hook.LastEntry().Data["kind"]).To(Equal(event.ChangesetAlistList))
			Expect(hook.LastEntry().Data["alist_uuid"]).To(Equal(alistUUID))
			Expect(hook.LastEntry().Data["action"]).To(Equal(event.ActionDeleted))
		})

		It("created or updated", func() {
			hugoHelper.OnEvent(moment)
			Expect(hook.LastEntry().Data["kind"]).To(Equal(event.ChangesetAlistList))
			Expect(hook.LastEntry().Data["alist_uuid"]).To(Equal(alistUUID))
			Expect(hook.LastEntry().Data["action"]).To(Equal(event.ActionCreated))
		})
	})

	When("User has been deleted or update to the users lists", func() {
		var (
			userUUID string
		)
		BeforeEach(func() {
			userUUID = "fake-user-123"
			moment = event.Eventlog{
				Kind:   event.ChangesetAlistUser,
				UUID:   userUUID,
				Data:   make([]alist.ShortInfo, 0),
				Action: event.ActionUpdated,
			}
		})

		It("Deleted", func() {
			moment.Action = event.ActionDeleted
			hugoHelper.OnEvent(moment)
			Expect(hook.LastEntry().Data["kind"]).To(Equal(event.ChangesetAlistUser))
			Expect(hook.LastEntry().Data["user_uuid"]).To(Equal(userUUID))
			Expect(hook.LastEntry().Data["action"]).To(Equal(event.ActionDeleted))
		})

		It("update the users lists", func() {
			hugoHelper.OnEvent(moment)
			Expect(hook.LastEntry().Data["kind"]).To(Equal(event.ChangesetAlistUser))
			Expect(hook.LastEntry().Data["user_uuid"]).To(Equal(userUUID))
			Expect(hook.LastEntry().Data["action"]).To(Equal(event.ActionUpdated))
			Expect(hook.LastEntry().Data["total_lists"]).To(Equal(0))
		})
	})

	When("Public lists are updated", func() {
		It("upsert of the public lists", func() {
			moment = event.Eventlog{
				Kind:   event.ChangesetAlistPublic,
				Data:   make([]alist.ShortInfo, 0),
				Action: event.ActionUpdated,
			}

			hugoHelper.OnEvent(moment)
			Expect(hook.LastEntry().Data["kind"]).To(Equal(event.ChangesetAlistPublic))
			Expect(hook.LastEntry().Data["action"]).To(Equal(event.ActionUpdated))
			Expect(hook.LastEntry().Data["total_lists"]).To(Equal(0))
		})
	})

	When("A challenge has been created, updated or deleted", func() {
		var (
			challengeUUID string
		)
		BeforeEach(func() {
			challengeUUID = "fake-challenge-123"
			moment = event.Eventlog{
				Kind: event.ChangesetChallenge,
				UUID: challengeUUID,
				Data: challenge.ChallengeInfo{
					UUID: challengeUUID,
				},
				Action: event.ActionUpdated,
			}
		})

		It("Deleted", func() {
			moment.Action = event.ActionDeleted
			hugoHelper.OnEvent(moment)
			Expect(hook.LastEntry().Data["kind"]).To(Equal(event.ChangesetChallenge))
			Expect(hook.LastEntry().Data["challenge_uuid"]).To(Equal(challengeUUID))
			Expect(hook.LastEntry().Data["action"]).To(Equal(event.ActionDeleted))
		})

		It("Upsert", func() {
			hugoHelper.OnEvent(moment)
			Expect(hook.LastEntry().Data["kind"]).To(Equal(event.ChangesetChallenge))
			Expect(hook.LastEntry().Data["challenge_uuid"]).To(Equal(challengeUUID))
			Expect(hook.LastEntry().Data["action"]).To(Equal(event.ActionUpdated))
		})
	})
})

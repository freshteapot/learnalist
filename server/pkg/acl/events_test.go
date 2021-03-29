package acl_test

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
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

		want    error
		service acl.AclService
		aclRepo *mocks.Acl

		userUUID string
		moment   event.Eventlog
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()

		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		eventMessageBus.On("Subscribe", event.TopicMonolog, "aclService", mock.Anything)

		aclRepo = &mocks.Acl{}

		want = errors.New("want")
		userUUID = "fake-user-123"
	})

	When("New user registered", func() {
		BeforeEach(func() {
			moment = event.Eventlog{
				Kind: event.ApiUserRegister,
				Data: event.EventNewUser{
					UUID: userUUID,
					Kind: event.KindUserRegisterUsername,
					Data: user.UserPreference{
						Acl: user.ACL{
							PublicListWrite: 1,
						},
					},
				},
			}
		})

		It("Issue saving", func() {
			testutils.SetLoggerToPanicOnFatal(logger)
			aclRepo.On("GrantUserPublicListWriteAccess", userUUID).Return(want)
			service = acl.NewService(aclRepo, logger)
			Expect(func() { service.OnEvent(moment) }).Should(Panic())
			Expect(hook.LastEntry().Data["error"]).To(Equal(want))
		})

		It("Success", func() {
			aclRepo.On("GrantUserPublicListWriteAccess", userUUID).Return(nil)
			service = acl.NewService(aclRepo, logger)
			service.OnEvent(moment)
			Expect(hook.LastEntry()).To(BeNil())
		})
	})

	When("Changing users right to create public lists", func() {
		When("Granting", func() {
			BeforeEach(func() {
				moment = event.Eventlog{
					Kind: acl.EventPublicListAccess,
					Data: acl.EventPublicListAccessData{
						UserUUID: userUUID,
						Action:   aclKeys.ActionGrant,
					},
					TriggeredBy: "cmd",
				}
			})

			It("Issue saving", func() {
				testutils.SetLoggerToPanicOnFatal(logger)
				aclRepo.On("GrantUserPublicListWriteAccess", userUUID).Return(want)
				service = acl.NewService(aclRepo, logger)
				Expect(func() { service.OnEvent(moment) }).Should(Panic())
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			})

			It("Success", func() {
				aclRepo.On("GrantUserPublicListWriteAccess", userUUID).Return(nil)
				service = acl.NewService(aclRepo, logger)
				service.OnEvent(moment)
				Expect(hook.LastEntry().Message).To(Equal("Access granted"))
			})
		})

		When("Revoking", func() {
			BeforeEach(func() {
				moment = event.Eventlog{
					Kind: acl.EventPublicListAccess,
					Data: acl.EventPublicListAccessData{
						UserUUID: userUUID,
						Action:   aclKeys.ActionRevoke,
					},
					TriggeredBy: "cmd",
				}
			})

			It("Issue saving", func() {
				testutils.SetLoggerToPanicOnFatal(logger)
				aclRepo.On("RevokeUserPublicListWriteAccess", userUUID).Return(want)
				service = acl.NewService(aclRepo, logger)
				Expect(func() { service.OnEvent(moment) }).Should(Panic())
				Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			})

			It("Success", func() {
				aclRepo.On("RevokeUserPublicListWriteAccess", userUUID).Return(nil)
				service = acl.NewService(aclRepo, logger)
				service.OnEvent(moment)
				Expect(hook.LastEntry().Message).To(Equal("Access revoked"))
			})
		})
	})
})

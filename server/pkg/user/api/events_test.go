package api_test

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userApi "github.com/freshteapot/learnalist-api/server/pkg/user/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing User API", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		hook            *test.Hook

		service userApi.UserService

		oauthHandlers               oauth.Handlers
		userManagement              *mocks.Management
		userFromIDP                 *mocks.UserFromIDP
		userSession                 *mocks.Session
		userInfoRepo                *mocks.UserInfoRepository
		userWithUsernameAndPassword *mocks.UserWithUsernameAndPassword
		oauthApple                  *mocks.OAuth2ConfigInterface
		oauthGoogle                 *mocks.OAuth2ConfigInterface
		aclRepo                     *mocks.Acl

		want     error
		userUUID string
		moment   event.Eventlog
		pref     user.UserPreference
	)

	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		eventMessageBus.On("Subscribe", event.TopicMonolog, "userService", mock.Anything)

		want = errors.New("want")
		userUUID = "fake-user-123"

		oauthHandlers = oauth.Handlers{}
		oauthGoogle = &mocks.OAuth2ConfigInterface{}
		oauthApple = &mocks.OAuth2ConfigInterface{}
		oauthHandlers.AddAppleID(oauthApple)
		oauthHandlers.AddGoogle(oauthGoogle)

		aclRepo = &mocks.Acl{}
		userManagement = &mocks.Management{}
		userSession = &mocks.Session{}
		userFromIDP = &mocks.UserFromIDP{}
		userInfoRepo = &mocks.UserInfoRepository{}
		userWithUsernameAndPassword = &mocks.UserWithUsernameAndPassword{}
		service = userApi.NewService(
			oauthHandlers,
			aclRepo,
			userManagement,
			userFromIDP,
			userSession,
			userWithUsernameAndPassword,
			userInfoRepo,
			"",
			logger)

	})

	When("Handling user access to public lists", func() {
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

		It("Bad data in the event", func() {
			moment.Data = `{`
			service.OnEvent(moment)

		})

		It("Failed to get info due to storage issues", func() {
			testutils.SetLoggerToPanicOnFatal(logger)

			userInfoRepo.On("Get", userUUID).Return(pref, want)
			Expect(func() {
				service.OnEvent(moment)
			}).Should(Panic())
			Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo)
		})

		It("failed to save info", func() {
			testutils.SetLoggerToPanicOnFatal(logger)

			userInfoRepo.On("Get", userUUID).Return(pref, nil)
			userInfoRepo.On("Save", userUUID, mock.MatchedBy(func(expectPref user.UserPreference) bool {
				Expect(expectPref.Acl.PublicListWrite).To(Equal(1))
				return true
			})).Return(want)

			Expect(func() {
				service.OnEvent(moment)
			}).Should(Panic())
			Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo)
		})

		It("Successfully updated the users info", func() {

			tests := []struct {
				Action       string
				PublicAccess int
			}{
				{
					Action:       aclKeys.ActionGrant,
					PublicAccess: 1,
				},
				{
					Action:       aclKeys.ActionRevoke,
					PublicAccess: 0,
				},
			}

			for _, test := range tests {
				moment.Data = acl.EventPublicListAccessData{
					UserUUID: userUUID,
					Action:   test.Action,
				}

				userInfoRepo.On("Get", userUUID).Return(pref, nil).Once()
				userInfoRepo.On("Save", userUUID, mock.MatchedBy(func(expectPref user.UserPreference) bool {

					Expect(expectPref.Acl.PublicListWrite).To(Equal(test.PublicAccess))
					return true
				})).Return(nil).Once()

				service.OnEvent(moment)
			}
		})
	})

	When("A new user is registered", func() {
		BeforeEach(func() {
			moment = event.Eventlog{
				Kind: event.ApiUserRegister,
				Data: event.EventNewUser{
					UUID: userUUID,
					Kind: event.KindUserRegisterUsername,
					Data: pref,
				},
			}
		})

		It("Issue saving info to repo", func() {
			testutils.SetLoggerToPanicOnFatal(logger)

			userInfoRepo.On("Save", userUUID, pref).Return(want)
			Expect(func() {
				service.OnEvent(moment)
			}).Should(Panic())
			Expect(hook.LastEntry().Data["error"]).To(Equal(want))
			mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo)
		})

		It("Saved", func() {
			userInfoRepo.On("Save", userUUID, pref).Return(nil)
			service.OnEvent(moment)
			userInfoRepo.AssertExpectations(GinkgoT())
		})
	})

	When("Adding over time is finished or removed", func() {
		var (
			userUUID, alistUUID, dripfeedUUID string
		)
		BeforeEach(func() {
			userUUID = "fake-user-123"
			alistUUID = "fake-list-123"
			dripfeedUUID = "fake-dripfeed-123"
			moment = event.Eventlog{
				Kind: dripfeed.EventDripfeedFinished,
				Data: openapi.SpacedRepetitionOvertimeInfo{
					DripfeedUuid: dripfeedUUID,
					UserUuid:     userUUID,
					AlistUuid:    alistUUID,
				},
				UUID: dripfeedUUID,
			}
			pref = user.UserPreference{}
		})

		It("Failed to write to repo", func() {
			testutils.SetLoggerToPanicOnFatal(logger)

			userInfoRepo.On("Get", userUUID).Return(pref, want)

			Expect(func() {
				service.OnEvent(moment)
			}).Should(Panic())
			Expect(hook.LastEntry().Data["error"]).To(Equal(want))

			mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo)
		})

		It("Success", func() {
			pref := user.UserPreference{
				SpacedRepetition: &user.SpacedRepetition{
					ListsOvertime: []string{alistUUID},
				},
			}
			userInfoRepo.On("Get", userUUID).Return(pref, nil)
			userInfoRepo.On("Save", userUUID, mock.MatchedBy(func(expectPref user.UserPreference) bool {
				Expect(expectPref.SpacedRepetition.ListsOvertime).To(Equal([]string{}))
				return true
			})).Return(nil)

			service.OnEvent(moment)
			mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo)
		})
	})

	When("An item is added for learning over time", func() {
		var (
			userUUID, alistUUID, dripfeedUUID string
		)
		BeforeEach(func() {
			userUUID = "fake-user-123"
			alistUUID = "fake-list-123"
			dripfeedUUID = "fake-dripfeed-123"
			moment = event.Eventlog{
				Kind: dripfeed.EventDripfeedAdded,
				Data: openapi.SpacedRepetitionOvertimeInfo{
					DripfeedUuid: dripfeedUUID,
					UserUuid:     userUUID,
					AlistUuid:    alistUUID,
				},
				UUID: dripfeedUUID,
			}
			pref = user.UserPreference{}
		})

		It("Failed to write to repo", func() {
			testutils.SetLoggerToPanicOnFatal(logger)

			userInfoRepo.On("Get", userUUID).Return(pref, want)

			Expect(func() {
				service.OnEvent(moment)
			}).Should(Panic())
			Expect(hook.LastEntry().Data["error"]).To(Equal(want))

			mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo)
		})

		It("Success", func() {
			pref := user.UserPreference{}
			userInfoRepo.On("Get", userUUID).Return(pref, nil)
			userInfoRepo.On("Save", userUUID, mock.MatchedBy(func(expectPref user.UserPreference) bool {
				Expect(expectPref.SpacedRepetition.ListsOvertime).To(Equal([]string{alistUUID}))
				return true
			})).Return(nil)

			service.OnEvent(moment)
			mock.AssertExpectationsForObjects(GinkgoT(), userInfoRepo)
		})
	})
})

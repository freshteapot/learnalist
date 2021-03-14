package api_test

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userApi "github.com/freshteapot/learnalist-api/server/pkg/user/api"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Spaced Repetition user info helpers", func() {
	var (
		userInfoRepo *mocks.UserInfoRepository
	)

	BeforeEach(func() {
		userInfoRepo = &mocks.UserInfoRepository{}
	})

	When("Saving", func() {
		It("Issue getting data", func() {
			want := errors.New("want")
			userInfoRepo.On("Get", "fake-chris").Return(user.UserPreference{}, want)
			err := userApi.AppendAndSaveSpacedRepetition(userInfoRepo, "fake-chris", "123")
			Expect(err).To(Equal(want))
		})

		It("Already on the list", func() {
			userInfoRepo.On("Get", "fake-chris").Return(user.UserPreference{
				SpacedRepetition: &user.SpacedRepetition{
					ListsOvertime: []string{"123"},
				},
			}, nil)
			err := userApi.AppendAndSaveSpacedRepetition(userInfoRepo, "fake-chris", "123")
			Expect(err).NotTo(HaveOccurred())
		})

		It("Add to list", func() {
			pref := user.UserPreference{}
			pref.SpacedRepetition = &user.SpacedRepetition{
				ListsOvertime: []string{"123"},
			}

			userInfoRepo.On("Get", "fake-chris").Return(user.UserPreference{}, utils.ErrNotFound)
			userInfoRepo.On("Save", "fake-chris", pref).Return(nil)
			err := userApi.AppendAndSaveSpacedRepetition(userInfoRepo, "fake-chris", "123")
			Expect(err).To(BeNil())
		})
	})

	When("Removing", func() {
		It("Issue getting data", func() {
			want := errors.New("want")
			userInfoRepo.On("Get", "fake-chris").Return(user.UserPreference{}, want)
			err := userApi.RemoveAndSaveSpacedRepetition(userInfoRepo, "fake-chris", "123")
			Expect(err).To(Equal(want))
		})

		It("On the list, remove it", func() {
			pref := user.UserPreference{}
			pref.SpacedRepetition = &user.SpacedRepetition{
				ListsOvertime: []string{},
			}
			userInfoRepo.On("Get", "fake-chris").Return(user.UserPreference{
				SpacedRepetition: &user.SpacedRepetition{
					ListsOvertime: []string{"123"},
				},
			}, nil)
			userInfoRepo.On("Save", "fake-chris", pref).Return(nil)
			err := userApi.RemoveAndSaveSpacedRepetition(userInfoRepo, "fake-chris", "123")
			Expect(err).NotTo(HaveOccurred())
		})

		It("Not on the list", func() {
			userInfoRepo.On("Get", "fake-chris").Return(user.UserPreference{
				SpacedRepetition: &user.SpacedRepetition{
					ListsOvertime: []string{"123"},
				},
			}, nil)
			err := userApi.RemoveAndSaveSpacedRepetition(userInfoRepo, "fake-chris", "456")
			Expect(err).To(BeNil())
		})

		It("No user info", func() {
			userInfoRepo.On("Get", "fake-chris").Return(user.UserPreference{}, utils.ErrNotFound)
			err := userApi.RemoveAndSaveSpacedRepetition(userInfoRepo, "fake-chris", "456")
			Expect(err).To(BeNil())
		})

		It("Bad json in the db, should never happen", func() {
			userInfoRepo.On("Get", "fake-chris").Return(user.UserPreference{}, nil)
			err := userApi.RemoveAndSaveSpacedRepetition(userInfoRepo, "fake-chris", "456")
			Expect(err).To(BeNil())
		})
	})
})

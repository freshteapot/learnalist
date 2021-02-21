package info_test

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/user/info"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Spaced Repetition user info helpers", func() {
	var (
		userManagementRepo *mocks.ManagementStorage
	)

	BeforeEach(func() {
		userManagementRepo = &mocks.ManagementStorage{}
	})

	When("Saving", func() {
		It("Issue getting data", func() {
			want := errors.New("want")
			userManagementRepo.On("GetInfo", "fake-chris").Return([]byte(`{}`), want)
			err := info.AppendAndSaveSpacedRepetition(userManagementRepo, "fake-chris", "123")
			Expect(err).To(Equal(want))
		})

		It("Already on the list", func() {
			userManagementRepo.On("GetInfo", "fake-chris").Return([]byte(`{"spaced_repetition":{"lists_overtime":["123"]}}`), nil)
			err := info.AppendAndSaveSpacedRepetition(userManagementRepo, "fake-chris", "123")
			Expect(err).NotTo(HaveOccurred())
		})

		It("Add to list", func() {
			pref := user.UserPreference{}
			pref.SpacedRepetition = &user.SpacedRepetition{
				ListsOvertime: []string{"123"},
			}
			b, _ := json.Marshal(pref)

			userManagementRepo.On("GetInfo", "fake-chris").Return([]byte(`{}`), utils.ErrNotFound)
			userManagementRepo.On("SaveInfo", "fake-chris", b).Return(nil)
			err := info.AppendAndSaveSpacedRepetition(userManagementRepo, "fake-chris", "123")
			Expect(err).To(BeNil())
		})
	})

	When("Removing", func() {
		It("Issue getting data", func() {
			want := errors.New("want")
			userManagementRepo.On("GetInfo", "fake-chris").Return([]byte(`{}`), want)
			err := info.RemoveAndSaveSpacedRepetition(userManagementRepo, "fake-chris", "123")
			Expect(err).To(Equal(want))
		})

		It("On the list, remove it", func() {
			pref := user.UserPreference{}
			pref.SpacedRepetition = &user.SpacedRepetition{
				ListsOvertime: []string{},
			}
			b, _ := json.Marshal(pref)
			userManagementRepo.On("GetInfo", "fake-chris").Return([]byte(`{"spaced_repetition":{"lists_overtime":["123"]}}`), nil)
			userManagementRepo.On("SaveInfo", "fake-chris", b).Return(nil)
			err := info.RemoveAndSaveSpacedRepetition(userManagementRepo, "fake-chris", "123")
			Expect(err).NotTo(HaveOccurred())
		})

		It("Not on the list", func() {
			userManagementRepo.On("GetInfo", "fake-chris").Return([]byte(`{"spaced_repetition":{"lists_overtime":["123"]}}`), nil)
			err := info.RemoveAndSaveSpacedRepetition(userManagementRepo, "fake-chris", "456")
			Expect(err).To(BeNil())
		})

		It("No user info", func() {
			userManagementRepo.On("GetInfo", "fake-chris").Return([]byte(``), utils.ErrNotFound)
			err := info.RemoveAndSaveSpacedRepetition(userManagementRepo, "fake-chris", "456")
			Expect(err).To(BeNil())
		})

		It("Bad json in the db, should never happen", func() {
			userManagementRepo.On("GetInfo", "fake-chris").Return([]byte(`{`), nil)
			err := info.RemoveAndSaveSpacedRepetition(userManagementRepo, "fake-chris", "456")
			Expect(err).To(BeNil())
		})

		It("Bad json in the db, should never happen", func() {
			userManagementRepo.On("GetInfo", "fake-chris").Return([]byte(`{}`), nil)
			err := info.RemoveAndSaveSpacedRepetition(userManagementRepo, "fake-chris", "456")
			Expect(err).To(BeNil())
		})
	})
})

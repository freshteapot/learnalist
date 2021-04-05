package models_test

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing from present", func() {
	var (
		dal          *models.DAL
		userUUID     string
		user         *uuid.User
		labels       *mocks.LabelReadWriter
		acl          *mocks.Acl
		aListStorage *mocks.DatastoreAlists
	)

	BeforeEach(func() {
		acl = &mocks.Acl{}
		labels = &mocks.LabelReadWriter{}
		aListStorage = &mocks.DatastoreAlists{}
		dal = models.NewDAL(acl, aListStorage, labels)
	})

	When("Testing info.from is present", func() {
		It("Do not let the from object be modified", func() {
			userUUID = "fake-user-123"
			user = &uuid.User{
				Uuid: userUUID,
			}

			aList := alist.NewTypeV1()
			aList.Uuid = "fake-list-123"
			aList.Info.Title = "A title"
			aList.Info.SharedWith = aclKeys.NotShared
			aList.User = *user
			aList.Info.From = &openapi.AlistFrom{}
			aList.Info.From.Kind = "quizlet"
			aList.Info.From.RefUrl = "https://quizlet.com/xxx"
			aList.Info.From.ExtUuid = "xxx"

			currentAlist := aList
			currentAlist.Info.SharedWith = aclKeys.NotShared
			currentAlist.Info.From = &openapi.AlistFrom{}
			currentAlist.Info.From.Kind = "quizlet"
			currentAlist.Info.From.RefUrl = "https://quizlet.com/xxx"
			currentAlist.Info.From.ExtUuid = ""

			aListStorage.On("GetAlist", aList.Uuid).Return(currentAlist, nil)
			_, err := dal.SaveAlist(http.MethodPut, aList)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(i18n.ErrorInputSaveAlistOperationFromModify))
		})

		It("If learnalist, let the shared attribute be changed", func() {
			userUUID = "fake-user-123"
			user = &uuid.User{
				Uuid: userUUID,
			}

			aList := alist.NewTypeV1()
			aList.Uuid = "fake-list-123"
			aList.Info.Title = "A title"
			aList.Info.SharedWith = aclKeys.SharedWithFriends
			aList.User = *user
			aList.Info.From = &openapi.AlistFrom{}
			aList.Info.From.Kind = "learnalist"
			aList.Info.From.RefUrl = "https://learnalist.net/xxx"
			aList.Info.From.ExtUuid = "xxx"

			currentAlist := aList
			currentAlist.Info.SharedWith = aclKeys.NotShared
			aListStorage.On("GetAlist", aList.Uuid).Return(currentAlist, nil)
			aListStorage.On("SaveAlist", http.MethodPut, aList).Return(currentAlist, nil)
			labels.On("RemoveLabelsForAlist", aList.Uuid).Return(nil)
			acl.On("ShareListWithFriends", aList.Uuid).Return(nil)
			_, err := dal.SaveAlist(http.MethodPut, aList)
			Expect(err).To(BeNil())
		})
	})
})

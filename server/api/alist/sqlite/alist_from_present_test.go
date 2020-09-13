package sqlite_test

import (
	"errors"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	storage "github.com/freshteapot/learnalist-api/server/api/alist/sqlite"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Alist", func() {
	var (
		dbCon        *sqlx.DB
		mockSql      sqlmock.Sqlmock
		aListStorage alist.DatastoreAlists
		logger       logrus.FieldLogger
	)

	BeforeEach(func() {
		logger, _ = test.NewNullLogger()
		dbCon, mockSql, _ = helper.GetMockDB()
		aListStorage = storage.NewAlist(dbCon, logger)
	})

	AfterEach(func() {
		dbCon.Close()
	})

	It("", func() {
		fmt.Println(mockSql, aListStorage)
		Expect("").To(Equal(""))
	})

	When("GetAllListsByUser", func() {
		var (
			userUUID = "fake-user-123"
		)

		Specify("Sucessfully removed", func() {
			uuid := "fake-list-123"
			title := "I am a title"
			rs := sqlmock.NewRows([]string{"title", "uuid"}).
				AddRow(
					title,
					uuid,
				)

			mockSql.ExpectQuery(storage.SqlGetAllListsByUser).
				WillReturnRows(rs)

			resp := aListStorage.GetAllListsByUser(userUUID)
			Expect(len(resp)).To(Equal(1))
			Expect(resp[0].UUID).To(Equal(uuid))
			Expect(resp[0].Title).To(Equal(title))
		})
	})

	When("GetPublicLists", func() {
		Specify("Sucessfully removed", func() {
			uuid := "fake-list-123"
			title := "I am a title"
			rs := sqlmock.NewRows([]string{"uuid", "title"}).
				AddRow(
					uuid,
					title,
				)

			mockSql.ExpectQuery(storage.SqlGetPublicLists).
				WillReturnRows(rs)

			resp := aListStorage.GetPublicLists()
			Expect(len(resp)).To(Equal(1))
			Expect(resp[0].UUID).To(Equal(uuid))
			Expect(resp[0].Title).To(Equal(title))
		})
	})

	When("Removing a list", func() {
		var (
			alistUUID = "fake-list-123"
			userUUID  = "fake-user-123"
		)

		Specify("Sucessfully removed", func() {
			mockSql.ExpectExec(storage.SqlDeleteItemByUserAndUUID).
				WithArgs(alistUUID, userUUID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := aListStorage.RemoveAlist(alistUUID, userUUID)
			Expect(err).To(BeNil())
		})

		Specify("An error occurred", func() {
			want := errors.New("Fail")
			mockSql.ExpectExec(storage.SqlDeleteItemByUserAndUUID).
				WithArgs(alistUUID, userUUID).
				WillReturnError(want)

			err := aListStorage.RemoveAlist(alistUUID, userUUID)
			Expect(err).Should(HaveOccurred())
			Expect(err).To(Equal(want))
		})
	})
})

/*
import (
	"encoding/json"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	alistStorage "github.com/freshteapot/learnalist-api/server/api/alist/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Models with sqlmock", func() {
	var (
		dal      *models.DAL
		dbCon    *sqlx.DB
		mockSql  sqlmock.Sqlmock
		userUUID string
		user     *uuid.User
		labels   *mocks.LabelReadWriter
		acl      *mocks.Acl
	)

	BeforeEach(func() {
		dbCon, mockSql, _ = helper.GetMockDB()
		acl = &mocks.Acl{}
		userSession := &mocks.Session{}
		userFromIDP := &mocks.UserFromIDP{}
		userWithUsernameAndPassword := &mocks.UserWithUsernameAndPassword{}
		oauthHandler := &mocks.OAuthReadWriter{}
		labels = &mocks.LabelReadWriter{}
		dal = models.NewDAL(dbCon, acl, labels, userSession, userFromIDP, userWithUsernameAndPassword, oauthHandler)
	})

	AfterEach(func() {
		dbCon.Close()
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

			b, _ := json.Marshal(currentAlist)

			rs := sqlmock.NewRows([]string{"uuid", "body", "user_uuid", "list_type"}).
				AddRow(
					aList.Uuid,
					string(b),
					aList.User.Uuid,
					aList.Info.ListType,
				)

			mockSql.ExpectQuery(alistStorage.SQL_GET_ITEM_BY_UUID).
				WithArgs(aList.Uuid).
				WillReturnRows(rs)

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

			bNew, _ := json.Marshal(aList)
			b, _ := json.Marshal(currentAlist)

			rs := sqlmock.NewRows([]string{"uuid", "body", "user_uuid", "list_type"}).
				AddRow(
					aList.Uuid,
					string(b),
					aList.User.Uuid,
					aList.Info.ListType,
				)

			mockSql.ExpectQuery(alistStorage.SQL_GET_ITEM_BY_UUID).
				WithArgs(aList.Uuid).
				WillReturnRows(rs)

			mockSql.ExpectExec(alistStorage.SQL_UPDATE_LIST).
				WithArgs(aList.Info.ListType, string(bNew), aList.User.Uuid, aList.Uuid).
				WillReturnResult(sqlmock.NewResult(1, 1))

			labels.On("RemoveLabelsForAlist", aList.Uuid).Return(nil)
			acl.On("ShareListWithFriends", aList.Uuid).Return(nil)
			_, err := dal.SaveAlist(http.MethodPut, aList)
			Expect(err).To(BeNil())
		})
	})
})
*/

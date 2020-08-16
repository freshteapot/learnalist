package models_test

import (
	"encoding/json"
	"fmt"
	"net/http"

	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Models with sqlmock", func() {
	var (
		dal      *models.DAL
		err      error
		dbCon    *sqlx.DB
		mockSql  sqlmock.Sqlmock
		userUUID string
		user     *uuid.User
	)

	BeforeEach(func() {
		dbCon, mockSql, err = helper.GetMockDB()
		acl := &mocks.Acl{}
		userSession := &mocks.Session{}
		userFromIDP := &mocks.UserFromIDP{}
		userWithUsernameAndPassword := &mocks.UserWithUsernameAndPassword{}
		oauthHandler := &mocks.OAuthReadWriter{}
		labels := &mocks.LabelReadWriter{}
		dal = models.NewDAL(dbCon, acl, labels, userSession, userFromIDP, userWithUsernameAndPassword, oauthHandler)
		fmt.Println(err, mockSql)
	})

	AfterEach(func() {
		dbCon.Close()
	})

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
		currentAlist.Info.From.ExtUuid = "xxx"

		b, _ := json.Marshal(currentAlist)

		rs := sqlmock.NewRows([]string{"uuid", "body", "user_uuid", "list_type"}).
			AddRow(
				aList.Uuid,
				string(b),
				aList.User.Uuid,
				aList.Info.ListType,
			)

		mockSql.ExpectQuery(models.SQL_GET_ITEM_BY_UUID).
			WithArgs(aList.Uuid).
			WillReturnRows(rs)

		_, err := dal.SaveAlist(http.MethodPut, aList)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(i18n.InputSaveAlistOperationFromModify))
	})
})

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

var _ = Describe("Testing Alist Sqlite Storage", func() {
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

		Specify("Successfully removed", func() {
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
		Specify("Successfully removed", func() {
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

		Specify("Successfully removed", func() {
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

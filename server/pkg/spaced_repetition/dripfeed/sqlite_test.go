package dripfeed_test

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Spaced Repetitiion Overtime Repository Sqlite", func() {
	var (
		dbCon                  *sqlx.DB
		mockSql                sqlmock.Sqlmock
		repo                   dripfeed.DripfeedRepository
		dripfeedUUID, userUUID string
		want                   error
	)

	BeforeEach(func() {
		dbCon, mockSql, _ = helper.GetMockDB()
		repo = dripfeed.NewSqliteRepository(dbCon)
		userUUID = "fake-user-123"

		want = errors.New("fail")
		fmt.Println(userUUID)
	})

	AfterEach(func() {
		dbCon.Close()
	})

	When("Exists", func() {
		var (
			sqlExpect *sqlmock.ExpectedQuery
		)

		BeforeEach(func() {
			dripfeedUUID = "fake-dripfeed-uuid"
			sqlExpect = mockSql.ExpectQuery(dripfeed.SqlDripfeedItemExists)
		})

		It("Issue", func() {
			sqlExpect.WillReturnError(want).WithArgs(
				dripfeedUUID,
			)
			exists, err := repo.Exists(dripfeedUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
			Expect(exists).To(BeTrue())
		})

		It("Not found", func() {
			sqlExpect.WillReturnError(sql.ErrNoRows).WithArgs(
				dripfeedUUID,
			)
			exists, err := repo.Exists(dripfeedUUID)
			Expect(err).To(BeNil())
			Expect(exists).To(BeFalse())
		})

		It("Exists", func() {
			rs := sqlmock.NewRows([]string{""}).AddRow("1")
			sqlExpect.WillReturnRows(rs).WithArgs(
				dripfeedUUID,
			)
			exists, err := repo.Exists(dripfeedUUID)
			Expect(err).To(BeNil())
			Expect(exists).To(BeTrue())
		})
	})

	When("Deleting by user", func() {
		When("Transaction fails", func() {
			It("Begin", func() {
				mockSql.ExpectBegin().WillReturnError(want)
				err := repo.DeleteByUser(userUUID)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			When("Rollback", func() {
				It("Fails on removing items", func() {
					mockSql.ExpectBegin()
					mockSql.ExpectExec(dripfeed.SqlDeleteDripfeedItemByUser).
						WithArgs(userUUID).
						WillReturnError(want)

					err := repo.DeleteByUser(userUUID)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(want))
				})

				It("Fails on removing info", func() {
					mockSql.ExpectBegin()
					mockSql.ExpectExec(dripfeed.SqlDeleteDripfeedItemByUser).
						WithArgs(userUUID).
						WillReturnResult(sqlmock.NewResult(1, 1))

					mockSql.ExpectExec(dripfeed.SqlDeleteInfoByUser).
						WithArgs(userUUID).
						WillReturnError(want)

					err := repo.DeleteByUser(userUUID)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(want))
				})
			})

			It("Commit", func() {
				mockSql.ExpectBegin()
				mockSql.ExpectExec(dripfeed.SqlDeleteDripfeedItemByUser).
					WithArgs(userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mockSql.ExpectExec(dripfeed.SqlDeleteInfoByUser).
					WithArgs(userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mockSql.ExpectCommit().WillReturnError(want)

				err := repo.DeleteByUser(userUUID)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})
		})
		Specify("Successfully removed", func() {
			mockSql.ExpectBegin()
			mockSql.ExpectExec(dripfeed.SqlDeleteDripfeedItemByUser).
				WithArgs(userUUID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mockSql.ExpectExec(dripfeed.SqlDeleteInfoByUser).
				WithArgs(userUUID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mockSql.ExpectCommit()

			err := repo.DeleteByUser(userUUID)
			Expect(err).To(BeNil())
		})
	})

	When("Deleting by dripfeed", func() {
		When("Transaction fails", func() {
			It("Begin", func() {
				mockSql.ExpectBegin().WillReturnError(want)
				err := repo.DeleteByUUIDAndUserUUID(dripfeedUUID, userUUID)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			When("Rollback", func() {
				It("Fails on removing items", func() {
					mockSql.ExpectBegin()
					mockSql.ExpectExec(dripfeed.SqlDeleteDripfeedItemByDripfeedUUIDAndUserUUID).
						WithArgs(dripfeedUUID, userUUID).
						WillReturnError(want)

					err := repo.DeleteByUUIDAndUserUUID(dripfeedUUID, userUUID)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(want))
				})

				It("Fails on removing info", func() {
					mockSql.ExpectBegin()
					mockSql.ExpectExec(dripfeed.SqlDeleteDripfeedItemByDripfeedUUIDAndUserUUID).
						WithArgs(dripfeedUUID, userUUID).
						WillReturnResult(sqlmock.NewResult(1, 1))

					mockSql.ExpectExec(dripfeed.SqlDeleteInfoByDripfeedUUIDAndUserUUID).
						WithArgs(dripfeedUUID, userUUID).
						WillReturnError(want)

					err := repo.DeleteByUUIDAndUserUUID(dripfeedUUID, userUUID)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(want))
				})
			})

			It("Commit", func() {
				mockSql.ExpectBegin()
				mockSql.ExpectExec(dripfeed.SqlDeleteDripfeedItemByDripfeedUUIDAndUserUUID).
					WithArgs(dripfeedUUID, userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mockSql.ExpectExec(dripfeed.SqlDeleteInfoByDripfeedUUIDAndUserUUID).
					WithArgs(dripfeedUUID, userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mockSql.ExpectCommit().WillReturnError(want)

				err := repo.DeleteByUUIDAndUserUUID(dripfeedUUID, userUUID)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})
		})
		Specify("Successfully removed", func() {
			mockSql.ExpectBegin()
			mockSql.ExpectExec(dripfeed.SqlDeleteDripfeedItemByDripfeedUUIDAndUserUUID).
				WithArgs(dripfeedUUID, userUUID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mockSql.ExpectExec(dripfeed.SqlDeleteInfoByDripfeedUUIDAndUserUUID).
				WithArgs(dripfeedUUID, userUUID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mockSql.ExpectCommit()

			err := repo.DeleteByUUIDAndUserUUID(dripfeedUUID, userUUID)
			Expect(err).To(BeNil())
		})
	})
})

package dripfeed_test

import (
	"database/sql"
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
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
			sqlExpect.
				WithArgs(dripfeedUUID).
				WillReturnError(want)
			exists, err := repo.Exists(dripfeedUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
			Expect(exists).To(BeTrue())
		})

		It("Not found", func() {
			sqlExpect.
				WithArgs(dripfeedUUID).
				WillReturnError(sql.ErrNoRows)
			exists, err := repo.Exists(dripfeedUUID)
			Expect(err).To(BeNil())
			Expect(exists).To(BeFalse())
		})

		It("Exists", func() {
			rs := sqlmock.NewRows([]string{""}).AddRow("1")
			sqlExpect.
				WithArgs(dripfeedUUID).
				WillReturnRows(rs)

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

	When("Deleting spaced repetition item from dripfeed", func() {
		var (
			srsUUID   string
			sqlExpect *sqlmock.ExpectedExec
		)

		BeforeEach(func() {
			srsUUID = "fake-srs-item-123"
			sqlExpect = mockSql.
				ExpectExec(dripfeed.SqlDeleteDripfeedItemByUserAndSRS).
				WithArgs(userUUID, srsUUID)
		})

		It("Fails", func() {
			sqlExpect.WillReturnError(want)
			err := repo.DeleteAllByUserUUIDAndSpacedRepetitionUUID(userUUID, srsUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("Success", func() {
			sqlExpect.WillReturnResult(sqlmock.NewResult(1, 1))
			err := repo.DeleteAllByUserUUIDAndSpacedRepetitionUUID(userUUID, srsUUID)
			Expect(err).To(BeNil())
		})
	})

	// TODO remove this
	When("Saving info", func() {
		var (
			input     openapi.SpacedRepetitionOvertimeInfo
			sqlExpect *sqlmock.ExpectedExec
		)

		BeforeEach(func() {
			input.AlistUuid = "fake-list-123"
			input.UserUuid = userUUID
			input.DripfeedUuid = dripfeedUUID
			sqlExpect = mockSql.
				ExpectExec(dripfeed.SqlSaveDripfeedInfo).
				WithArgs(input.DripfeedUuid, input.UserUuid, input.AlistUuid)
		})

		It("Fails", func() {
			sqlExpect.
				WillReturnError(want)
			err := repo.SaveInfo(input)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("Success", func() {
			sqlExpect.
				WillReturnResult(sqlmock.NewResult(1, 1))
			err := repo.SaveInfo(input)
			Expect(err).To(BeNil())
		})
	})

	When("Adding all", func() {
		// TODO when list is empty, should we even bother adding it?
		// I think we should respond 422
		When("List is empty", func() {

		})

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

})

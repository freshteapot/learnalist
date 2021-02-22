package dripfeed_test

import (
	"database/sql"
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"

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

	When("Adding all", func() {
		var (
			alistUUID   string
			srsItemUUID string
			items       []string
		)
		BeforeEach(func() {
			dripfeedUUID = "1222536f4d7febb5cceb00138fe6d9792ab55101"
			alistUUID = "311d3938-fe9f-5da4-a181-c79572108927"
			srsItemUUID = "9c05511a31375a8a278a75207331bb1714e69dd1"
			// This would be actual srs json objects
			items = []string{
				`{"show":"hello world","kind":"v1","uuid":"9c05511a31375a8a278a75207331bb1714e69dd1","data":"hello world","settings":{"level":"0","when_next":"2021-02-20T13:06:47Z","created":"2021-02-20T12:06:47Z","ext_id":"1222536f4d7febb5cceb00138fe6d9792ab55101"}}`,
			}
		})

		When("Transaction fails", func() {
			It("Begin", func() {
				mockSql.ExpectBegin().WillReturnError(want)
				err := repo.AddAll(dripfeedUUID, userUUID, alistUUID, items)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			When("Rollback", func() {
				It("Fails on saving info", func() {
					mockSql.ExpectBegin()
					mockSql.ExpectExec(dripfeed.SqlSaveDripfeedInfo).
						WithArgs(dripfeedUUID, userUUID, alistUUID).
						WillReturnError(want)

					err := repo.AddAll(dripfeedUUID, userUUID, alistUUID, items)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(want))
				})

				It("Fails on saving item", func() {
					mockSql.ExpectBegin()
					mockSql.ExpectExec(dripfeed.SqlSaveDripfeedInfo).
						WithArgs(dripfeedUUID, userUUID, alistUUID).
						WillReturnResult(sqlmock.NewResult(1, 1))

					mockSql.ExpectExec(dripfeed.SqlDripfeedItemAddItem).
						WithArgs(
							dripfeedUUID,
							srsItemUUID,
							userUUID,
							alistUUID,
							items[0],
							0).
						WillReturnError(want)

					err := repo.AddAll(dripfeedUUID, userUUID, alistUUID, items)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(want))
				})
			})

			It("Commit", func() {
				mockSql.ExpectBegin()
				mockSql.ExpectExec(dripfeed.SqlSaveDripfeedInfo).
					WithArgs(dripfeedUUID, userUUID, alistUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mockSql.ExpectExec(dripfeed.SqlDripfeedItemAddItem).
					WithArgs(
						dripfeedUUID,
						srsItemUUID,
						userUUID,
						alistUUID,
						items[0],
						0).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mockSql.ExpectCommit().WillReturnError(want)

				err := repo.AddAll(dripfeedUUID, userUUID, alistUUID, items)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})
		})
		Specify("Successfully removed", func() {
			mockSql.ExpectBegin()
			mockSql.ExpectExec(dripfeed.SqlSaveDripfeedInfo).
				WithArgs(dripfeedUUID, userUUID, alistUUID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mockSql.ExpectExec(dripfeed.SqlDripfeedItemAddItem).
				WithArgs(
					dripfeedUUID,
					srsItemUUID,
					userUUID,
					alistUUID,
					items[0],
					0).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mockSql.ExpectCommit()

			err := repo.AddAll(dripfeedUUID, userUUID, alistUUID, items)
			Expect(err).To(BeNil())
		})
	})

	When("GetNext", func() {
		var (
			alistUUID   string
			srsItemUUID string
			srsItemBody string
			sqlExpect   *sqlmock.ExpectedQuery
		)

		BeforeEach(func() {
			dripfeedUUID = "1222536f4d7febb5cceb00138fe6d9792ab55101"
			alistUUID = "311d3938-fe9f-5da4-a181-c79572108927"
			srsItemUUID = "9c05511a31375a8a278a75207331bb1714e69dd1"
			// This would be actual srs json objects

			srsItemBody = `{"show":"hello world","kind":"v1","uuid":"9c05511a31375a8a278a75207331bb1714e69dd1","data":"hello world","settings":{"level":"0","when_next":"2021-02-20T13:06:47Z","created":"2021-02-20T12:06:47Z","ext_id":"1222536f4d7febb5cceb00138fe6d9792ab55101"}}`

			sqlExpect = mockSql.ExpectQuery(dripfeed.SqlDripfeedItemGetNext).WithArgs(dripfeedUUID)
		})

		It("Not found", func() {
			sqlExpect.WillReturnError(sql.ErrNoRows)
			_, err := repo.GetNext(dripfeedUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(utils.ErrNotFound))
		})

		It("Issue talking to db", func() {
			sqlExpect.WillReturnError(want)
			_, err := repo.GetNext(dripfeedUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("Success", func() {
			rs := sqlmock.NewRows([]string{
				"dripfeed_uuid",
				"srs_uuid",
				"user_uuid",
				"alist_uuid",
				"body",
				"position",
				"kind"}).AddRow(
				dripfeedUUID,
				srsItemUUID,
				userUUID,
				alistUUID,
				srsItemBody,
				0,
				"v1",
			)
			sqlExpect.WillReturnRows(rs)
			next, err := repo.GetNext(dripfeedUUID)

			Expect(err).To(BeNil())
			Expect(next.DripfeedUUID).To(Equal(dripfeedUUID))
			Expect(next.AlistUUID).To(Equal(alistUUID))
			Expect(next.Position).To(Equal(0))
			Expect(next.SrsBody).To(Equal([]byte(srsItemBody)))
			Expect(next.SrsUUID).To(Equal(srsItemUUID))
			Expect(next.SrsKind).To(Equal("v1"))
			Expect(next.UserUUID).To(Equal(userUUID))
		})
	})

	When("GetInfo", func() {
		var (
			alistUUID string
			sqlExpect *sqlmock.ExpectedQuery
		)

		BeforeEach(func() {
			dripfeedUUID = "1222536f4d7febb5cceb00138fe6d9792ab55101"
			alistUUID = "311d3938-fe9f-5da4-a181-c79572108927"

			sqlExpect = mockSql.ExpectQuery(dripfeed.SqlGetDripfeedInfo).WithArgs(dripfeedUUID)
		})

		It("Not found", func() {
			sqlExpect.WillReturnError(sql.ErrNoRows)
			_, err := repo.GetInfo(dripfeedUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(utils.ErrNotFound))
		})

		It("Issue talking to db", func() {
			sqlExpect.WillReturnError(want)
			_, err := repo.GetInfo(dripfeedUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("Success", func() {
			rs := sqlmock.NewRows([]string{
				"dripfeed_uuid",
				"user_uuid",
				"alist_uuid",
			}).AddRow(
				dripfeedUUID,
				userUUID,
				alistUUID,
			)
			sqlExpect.WillReturnRows(rs)
			info, err := repo.GetInfo(dripfeedUUID)

			Expect(err).To(BeNil())
			Expect(info.DripfeedUuid).To(Equal(dripfeedUUID))
			Expect(info.AlistUuid).To(Equal(alistUUID))
			Expect(info.UserUuid).To(Equal(userUUID))
		})
	})
})

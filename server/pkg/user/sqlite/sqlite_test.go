package sqlite_test

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	storage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/jmoiron/sqlx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getMockDB() (*sqlx.DB, sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		Fail(fmt.Sprintf("an error '%s' was not expected when opening a stub database connection", err.Error()))
	}

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	return sqlxDB, mock, err
}

var _ = Describe("Testing User", func() {
	When("Working with the user session", func() {
		var (
			err        error
			repoistory user.Session
			dbCon      *sqlx.DB
			mockSql    sqlmock.Sqlmock
		)

		BeforeEach(func() {
			dbCon, mockSql, err = getMockDB()
		})

		AfterEach(func() {
			dbCon.Close()
		})

		Context("Create", func() {
			It("Failed to save", func() {
				want := sql.ErrNoRows
				mockSql.ExpectExec(storage.UserSessionInsertEntry).
					WillReturnError(want)

				repoistory = storage.NewUserSession(dbCon)
				_, err = repoistory.Create()
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(sql.ErrNoRows))
			})

			It("Success", func() {
				mockSql.ExpectExec(storage.UserSessionInsertEntry).
					WillReturnResult(sqlmock.NewResult(1, 1))

				repoistory = storage.NewUserSession(dbCon)
				_, err = repoistory.Create()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Activate", func() {
			var session user.UserSession

			BeforeEach(func() {
				session = user.UserSession{
					Token:     "i-am-a-token",
					UserUUID:  "i-am-a-user",
					Challenge: "i-am-a-challenge",
				}
			})

			It("Failed to save", func() {
				want := errors.New("sql: fake")
				mockSql.ExpectExec(storage.UserSessionUpdateEntry).
					WillReturnError(want)

				repoistory = storage.NewUserSession(dbCon)
				err = repoistory.Activate(session)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Failed on sql result", func() {
				want := errors.New("sql: fake")
				mockSql.ExpectExec(storage.UserSessionUpdateEntry).
					WillReturnResult(sqlmock.NewErrorResult(want))

				repoistory = storage.NewUserSession(dbCon)
				err = repoistory.Activate(session)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Failed to find a record to update", func() {
				mockSql.ExpectExec(storage.UserSessionUpdateEntry).
					WillReturnResult(sqlmock.NewResult(0, 0))

				repoistory = storage.NewUserSession(dbCon)
				err = repoistory.Activate(session)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(i18n.ErrorUserSessionActivate))
			})

			It("Success", func() {
				mockSql.ExpectExec(storage.UserSessionUpdateEntry).
					WillReturnResult(sqlmock.NewResult(1, 1))

				repoistory = storage.NewUserSession(dbCon)
				err = repoistory.Activate(session)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})

package sqlite_test

import (
	"database/sql"
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	storage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/jmoiron/sqlx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing User", func() {
	When("Working with the user session", func() {
		var (
			err        error
			repository user.Session
			dbCon      *sqlx.DB
			mockSql    sqlmock.Sqlmock
		)

		BeforeEach(func() {
			dbCon, mockSql, err = helper.GetMockDB()
		})

		AfterEach(func() {
			dbCon.Close()
		})

		Context("Create", func() {
			It("Failed to save", func() {
				want := sql.ErrNoRows
				mockSql.ExpectExec(storage.UserSessionInsertEntry).
					WillReturnError(want)

				repository = storage.NewUserSession(dbCon)
				_, err = repository.CreateWithChallenge()
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(sql.ErrNoRows))
			})

			It("Success", func() {
				mockSql.ExpectExec(storage.UserSessionInsertEntry).
					WillReturnResult(sqlmock.NewResult(1, 1))

				repository = storage.NewUserSession(dbCon)
				_, err = repository.CreateWithChallenge()
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

				repository = storage.NewUserSession(dbCon)
				err = repository.Activate(session)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Failed on sql result", func() {
				want := errors.New("sql: fake")
				mockSql.ExpectExec(storage.UserSessionUpdateEntry).
					WillReturnResult(sqlmock.NewErrorResult(want))

				repository = storage.NewUserSession(dbCon)
				err = repository.Activate(session)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Failed to find a record to update", func() {
				mockSql.ExpectExec(storage.UserSessionUpdateEntry).
					WillReturnResult(sqlmock.NewResult(0, 0))

				repository = storage.NewUserSession(dbCon)
				err = repository.Activate(session)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(i18n.ErrorUserSessionActivate))
			})

			It("Success", func() {
				mockSql.ExpectExec(storage.UserSessionUpdateEntry).
					WillReturnResult(sqlmock.NewResult(1, 1))

				repository = storage.NewUserSession(dbCon)
				err = repository.Activate(session)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("GetUserUUIDByToken", func() {
			It("Trigger error", func() {
				token := "fake-token-123"
				want := errors.New("sql: fake")

				mockSql.ExpectQuery(storage.UserSessionSelectUserUUIDByToken).
					WillReturnError(want)

				repository = storage.NewUserSession(dbCon)
				_, err := repository.GetUserUUIDByToken(token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Find the user we are looking for", func() {
				token := "fake-token-123"
				userUUID := "fake-user-123"
				rs := sqlmock.NewRows([]string{
					"user_uuid",
				}).
					AddRow(userUUID)

				mockSql.ExpectQuery(storage.UserSessionSelectUserUUIDByToken).
					WithArgs(token).
					WillReturnRows(rs)

				repository = storage.NewUserSession(dbCon)
				found, err := repository.GetUserUUIDByToken(token)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(found).To(Equal(userUUID))
			})
		})

		Context("IsChallengeValid", func() {
			It("Error from db, not empty", func() {
				challenge := "challenge-123"
				want := errors.New("sql: fake")

				mockSql.ExpectQuery(storage.UserSessionSelectChallengeIsValid).
					WithArgs(challenge).
					WillReturnError(want)

				repository = storage.NewUserSession(dbCon)
				found, err := repository.IsChallengeValid(challenge)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
				Expect(found).To(BeFalse())
			})

			It("Challenge is not valid, no record was found", func() {
				challenge := "challenge-123"
				rs := sqlmock.NewRows([]string{""})
				mockSql.ExpectQuery(storage.UserSessionSelectChallengeIsValid).
					WithArgs(challenge).
					WillReturnRows(rs)

				repository = storage.NewUserSession(dbCon)
				found, err := repository.IsChallengeValid(challenge)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})

			It("Challenge is valid", func() {
				challenge := "challenge-123"
				rs := sqlmock.NewRows([]string{""}).AddRow("1")
				mockSql.ExpectQuery(storage.UserSessionSelectChallengeIsValid).
					WithArgs(challenge).
					WillReturnRows(rs)

				repository = storage.NewUserSession(dbCon)
				found, err := repository.IsChallengeValid(challenge)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(found).To(BeTrue())
			})
		})

		Context("RemoveSessionsForUser", func() {
			It("Trigger error", func() {
				userUUID := "fake-user-123"
				want := errors.New("sql: fake")

				mockSql.ExpectExec(storage.UserSessionDeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnError(want)

				repository = storage.NewUserSession(dbCon)
				err := repository.RemoveSessionsForUser(userUUID)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Remove all sessions", func() {
				userUUID := "fake-user-123"
				mockSql.ExpectExec(storage.UserSessionDeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				repository = storage.NewUserSession(dbCon)
				err := repository.RemoveSessionsForUser(userUUID)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("RemoveSessionForUser", func() {
			It("Trigger error", func() {
				userUUID := "fake-user-123"
				token := "fake-token-123"
				want := errors.New("sql: fake")

				mockSql.ExpectExec(storage.UserSessionDeleteByUserUUIDAndToken).
					WithArgs(userUUID, token).
					WillReturnError(want)

				repository = storage.NewUserSession(dbCon)
				err := repository.RemoveSessionForUser(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Remove specific session", func() {
				userUUID := "fake-user-123"
				token := "fake-token-123"

				mockSql.ExpectExec(storage.UserSessionDeleteByUserUUIDAndToken).
					WithArgs(userUUID, token).
					WillReturnResult(sqlmock.NewResult(1, 1))

				repository = storage.NewUserSession(dbCon)
				err := repository.RemoveSessionForUser(userUUID, token)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Remove unused sessions", func() {
			It("Trigger error", func() {
				want := errors.New("sql: fake")

				mockSql.ExpectExec(storage.UserSessionDeleteUnActiveChallenges).
					WillReturnError(want)

				repository = storage.NewUserSession(dbCon)
				err := repository.RemoveExpiredChallenges()
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Older than 12 hours has been cleaned", func() {
				mockSql.ExpectExec(storage.UserSessionDeleteUnActiveChallenges).
					WillReturnResult(sqlmock.NewResult(1, 1))

				repository = storage.NewUserSession(dbCon)
				err := repository.RemoveExpiredChallenges()
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Create a new session", func() {
			It("Trigger error", func() {
				userUUID := "fake-user-123"
				want := errors.New("sql: fake")

				mockSql.ExpectExec(storage.UserSessionInsertFullRecord).
					WillReturnError(want)

				repository = storage.NewUserSession(dbCon)
				_, err := repository.NewSession(userUUID)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("New session made", func() {
				userUUID := "fake-user-123"
				mockSql.ExpectExec(storage.UserSessionInsertFullRecord).
					WillReturnResult(sqlmock.NewResult(1, 1))

				repository = storage.NewUserSession(dbCon)
				session, err := repository.NewSession(userUUID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(session.UserUUID).To(Equal(userUUID))
			})
		})
	})
})

package sqlite_test

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	oauthSqlite "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"

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

var _ = Describe("Testing Oauth", func() {
	var (
		userUUID   string
		token      *oauth2.Token
		err        error
		repoistory oauth.OAuthReadWriter
		dbCon      *sqlx.DB
		mockSql    sqlmock.Sqlmock
	)

	BeforeEach(func() {
		dbCon, mockSql, err = getMockDB()
		token = new(oauth2.Token)
		token.AccessToken = "fake-access-token"
		token.TokenType = "Bearer"
		token.RefreshToken = "fake-refresh-token"
		token.Expiry = time.Now()
		userUUID = "fake-user-123"
	})

	AfterEach(func() {
		dbCon.Close()
	})

	When("Crud on token info", func() {
		Context("Getting a record", func() {
			It("When the user doesnt have any token info", func() {
				want := sql.ErrNoRows
				mockSql.ExpectQuery(oauthSqlite.SelectByUserUUID).
					WithArgs(userUUID).
					WillReturnError(want)

				repoistory = oauthSqlite.NewOAuthReadWriter(dbCon)
				_, err = repoistory.GetTokenInfo(userUUID)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(sql.ErrNoRows))
			})

			It("When the user doesnt have any token info", func() {
				rs := sqlmock.NewRows([]string{
					"user_uuid",
					"access_token",
					"token_type",
					"refresh_token",
					"expiry",
				}).
					AddRow(userUUID, token.AccessToken, token.TokenType, token.RefreshToken, token.Expiry)
				mockSql.ExpectQuery(oauthSqlite.SelectByUserUUID).
					WithArgs(userUUID).
					WillReturnRows(rs)

				repoistory = oauthSqlite.NewOAuthReadWriter(dbCon)
				found, err := repoistory.GetTokenInfo(userUUID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(found.AccessToken).To(Equal(token.AccessToken))
			})
		})

		Context("Writing token info", func() {
			It("Happy path", func() {
				mockSql.ExpectBegin()
				mockSql.ExpectExec(oauthSqlite.DeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectExec(oauthSqlite.InsertEntry).
					WithArgs(userUUID, token.AccessToken, token.TokenType, token.RefreshToken, token.Expiry).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectCommit()

				repoistory = oauthSqlite.NewOAuthReadWriter(dbCon)
				err := repoistory.WriteTokenInfo(userUUID, token)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("Fail on tx: begin", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin().WillReturnError(want)

				repoistory = oauthSqlite.NewOAuthReadWriter(dbCon)
				err := repoistory.WriteTokenInfo(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Fail on tx: deleting existing record", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin()
				mockSql.ExpectExec(oauthSqlite.DeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnError(want)

				repoistory = oauthSqlite.NewOAuthReadWriter(dbCon)
				err := repoistory.WriteTokenInfo(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Fail on tx: inserting token", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin()
				mockSql.ExpectExec(oauthSqlite.DeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectExec(oauthSqlite.InsertEntry).
					WillReturnError(want)

				repoistory = oauthSqlite.NewOAuthReadWriter(dbCon)
				err := repoistory.WriteTokenInfo(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Fail on tx: commit", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin()
				mockSql.ExpectExec(oauthSqlite.DeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectExec(oauthSqlite.InsertEntry).
					WithArgs(userUUID, token.AccessToken, token.TokenType, token.RefreshToken, token.Expiry).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectCommit().
					WillReturnError(want)

				repoistory = oauthSqlite.NewOAuthReadWriter(dbCon)
				err := repoistory.WriteTokenInfo(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})
		})
	})
})

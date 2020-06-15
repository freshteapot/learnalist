package sqlite_test

import (
	"database/sql"
	"errors"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Oauth", func() {
	var (
		userUUID   string
		token      *oauth2.Token
		err        error
		repository oauth.OAuthReadWriter
		dbCon      *sqlx.DB
		mockSql    sqlmock.Sqlmock
	)

	BeforeEach(func() {
		dbCon, mockSql, err = helper.GetMockDB()
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
				mockSql.ExpectQuery(oauthStorage.SelectByUserUUID).
					WithArgs(userUUID).
					WillReturnError(want)

				repository = oauthStorage.NewOAuthReadWriter(dbCon)
				_, err = repository.GetTokenInfo(userUUID)
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
					AddRow(userUUID, token.AccessToken, token.TokenType, token.RefreshToken, token.Expiry.UTC().Unix())
				mockSql.ExpectQuery(oauthStorage.SelectByUserUUID).
					WithArgs(userUUID).
					WillReturnRows(rs)

				repository = oauthStorage.NewOAuthReadWriter(dbCon)
				found, err := repository.GetTokenInfo(userUUID)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(found.AccessToken).To(Equal(token.AccessToken))
			})
		})

		Context("Writing token info", func() {
			It("Happy path", func() {
				mockSql.ExpectBegin()
				mockSql.ExpectExec(oauthStorage.DeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectExec(oauthStorage.InsertEntry).
					WithArgs(userUUID, token.AccessToken, token.TokenType, token.RefreshToken, token.Expiry.UTC().Unix()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectCommit()

				repository = oauthStorage.NewOAuthReadWriter(dbCon)
				err := repository.WriteTokenInfo(userUUID, token)
				Expect(err).ShouldNot(HaveOccurred())
			})

			It("Fail on tx: begin", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin().WillReturnError(want)

				repository = oauthStorage.NewOAuthReadWriter(dbCon)
				err := repository.WriteTokenInfo(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Fail on tx: deleting existing record", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin()
				mockSql.ExpectExec(oauthStorage.DeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnError(want)

				repository = oauthStorage.NewOAuthReadWriter(dbCon)
				err := repository.WriteTokenInfo(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Fail on tx: inserting token", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin()
				mockSql.ExpectExec(oauthStorage.DeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectExec(oauthStorage.InsertEntry).
					WillReturnError(want)

				repository = oauthStorage.NewOAuthReadWriter(dbCon)
				err := repository.WriteTokenInfo(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Fail on tx: commit", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin()
				mockSql.ExpectExec(oauthStorage.DeleteByUserUUID).
					WithArgs(userUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectExec(oauthStorage.InsertEntry).
					WithArgs(userUUID, token.AccessToken, token.TokenType, token.RefreshToken, token.Expiry.UTC().Unix()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectCommit().
					WillReturnError(want)

				repository = oauthStorage.NewOAuthReadWriter(dbCon)
				err := repository.WriteTokenInfo(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})
		})
	})
})

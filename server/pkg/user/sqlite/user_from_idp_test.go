package sqlite_test

import (
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	storage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/jmoiron/sqlx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing User from IDP", func() {
	When("Working with the user session", func() {
		var (
			err        error
			repository user.UserFromIDP
			dbCon      *sqlx.DB
			mockSql    sqlmock.Sqlmock
			idp        string
			kind       string
			identifier string
			info       []byte
		)

		BeforeEach(func() {
			dbCon, mockSql, err = helper.GetMockDB()
			idp = "google"
			identifier = "fake@learnalist.net"
			kind = "email"
			info = []byte(`{"hello": "world"}`)
		})

		AfterEach(func() {
			dbCon.Close()
		})

		Context("Register a new user", func() {
			It("Trigger error", func() {
				want := errors.New("sql: fake")
				mockSql.ExpectExec(storage.UserFromIDPInsertEntry).
					WillReturnError(want)

				repository = storage.NewUserFromIDP(dbCon)
				_, err = repository.Register(idp, identifier, info)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("New user registered", func() {
				mockSql.ExpectExec(storage.UserFromIDPInsertEntry).
					WillReturnResult(sqlmock.NewResult(1, 1))

				repository = storage.NewUserFromIDP(dbCon)
				_, err := repository.Register(idp, identifier, info)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("Lookup a user", func() {
			It("Trigger error", func() {
				want := errors.New("sql: fake")
				mockSql.ExpectQuery(storage.UserFromIDPFindUserUUID).
					WillReturnError(want)

				repository = storage.NewUserFromIDP(dbCon)
				_, err = repository.Lookup(idp, identifier, kind)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Nothing found", func() {
				want := user.ErrNotFound
				mockSql.ExpectQuery(storage.UserFromIDPFindUserUUID).
					WillReturnError(want)

				repository = storage.NewUserFromIDP(dbCon)
				_, err = repository.Lookup(idp, identifier, kind)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("Find a user", func() {
				userUUID := "fake-user-123"
				rs := sqlmock.NewRows([]string{"user_uuid"}).AddRow(userUUID)
				mockSql.ExpectQuery(storage.UserFromIDPFindUserUUID).
					WithArgs(idp, user.IDPKindEmail, identifier).
					WillReturnRows(rs)

				repository = storage.NewUserFromIDP(dbCon)
				found, err := repository.Lookup(idp, identifier, kind)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(found).To(Equal(userUUID))
			})
		})
	})
})

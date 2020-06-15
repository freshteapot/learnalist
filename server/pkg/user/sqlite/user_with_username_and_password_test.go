package sqlite_test

import (
	"database/sql"
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	storage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/jmoiron/sqlx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing User with username and password", func() {
	var (
		err        error
		repository user.UserWithUsernameAndPassword
		dbCon      *sqlx.DB
		mockSql    sqlmock.Sqlmock
		username   string
		hash       string
	)

	BeforeEach(func() {
		dbCon, mockSql, err = helper.GetMockDB()
		username = "iamusera"
		hash = "fake-hash"
		repository = storage.NewUserWithUsernameAndPassword(dbCon)
	})

	AfterEach(func() {
		dbCon.Close()
	})

	Context("Register a new user", func() {
		It("Trigger error", func() {
			want := errors.New("sql: fake")
			mockSql.ExpectExec(storage.UserWithUsernameAndPasswordInsertEntry).
				WillReturnError(want)

			_, err = repository.Register(username, hash)
			Expect(err).Should(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("Trigger error", func() {
			want := sql.ErrNoRows
			mockSql.ExpectExec(storage.UserWithUsernameAndPasswordInsertEntry).
				WillReturnError(want)

			_, err = repository.Register(username, hash)
			Expect(err).Should(HaveOccurred())
			Expect(err).To(Equal(want))
		})
	})

	Context("Lookup a user", func() {
		It("Trigger error", func() {
			want := errors.New("sql: fake")
			mockSql.ExpectQuery(storage.UserWithUsernameAndPasswordSelectUserUUIDByHash).
				WillReturnError(want)

			_, err = repository.Lookup(username, hash)
			Expect(err).Should(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("Nothing found", func() {
			want := sql.ErrNoRows
			mockSql.ExpectQuery(storage.UserWithUsernameAndPasswordSelectUserUUIDByHash).
				WithArgs(username, hash).
				WillReturnError(want)

			_, err = repository.Lookup(username, hash)
			Expect(err).Should(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("Find a user", func() {
			userUUID := "fake-user-123"
			rs := sqlmock.NewRows([]string{"uuid"}).AddRow(userUUID)
			mockSql.ExpectQuery(storage.UserWithUsernameAndPasswordSelectUserUUIDByHash).
				WithArgs(username, hash).
				WillReturnRows(rs)

			found, err := repository.Lookup(username, hash)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(found).To(Equal(userUUID))
		})
	})
})

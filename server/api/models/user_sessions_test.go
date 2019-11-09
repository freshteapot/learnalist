package models_test

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	aclSqlite "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/jmoiron/sqlx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing User Sessions", func() {
	var dal *models.DAL
	var db *sqlx.DB
	var acl acl.Acl
	BeforeEach(func() {
		db = database.NewTestDB()
		acl = aclSqlite.NewAcl(db)
		dal = models.NewDAL(db, acl)
	})

	AfterEach(func() {
		database.EmptyDatabase(db)
	})

	When("Crud", func() {
		It("Insert new session", func() {
			_, err := dal.InsertNewSession()
			Expect(err).ShouldNot(HaveOccurred())
		})

		When("Getting a session based on a token", func() {
			It("Successfully finds a token in the system", func() {
				token, _ := dal.InsertNewSession()
				session, err := dal.GetSessionByToken(token)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(session.UserUUID).To(Equal(models.SessionsWithoutUser))
			})

			It("Fails to find the session", func() {
				token := "fake"
				_, err := dal.GetSessionByToken(token)
				want := errors.New("sql: no rows in result set")
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})
		})

		When("Updating a session based on a token", func() {
			It("Successfully updaates a session linking user to token", func() {
				userUUID := "fake-123"
				token, _ := dal.InsertNewSession()
				err := dal.UpdateSession(userUUID, token)
				Expect(err).ShouldNot(HaveOccurred())
				session, err := dal.GetSessionByToken(token)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(session.UserUUID).To(Equal(userUUID))
			})

			It("Failed to update the session, linking", func() {
				want := errors.New("token not found, failing to link the user.")
				token := "fake"
				userUUID := "fake-123"
				err := dal.UpdateSession(userUUID, token)
				Expect(err).Should(HaveOccurred())
				Expect(err).To(Equal(want))
			})
		})
	})
})

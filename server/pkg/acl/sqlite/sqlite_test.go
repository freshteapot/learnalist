package sqlite_test

import (
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	aclSqlite "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/jmoiron/sqlx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Acl", func() {
	var db *sqlx.DB
	var doorKeeper acl.Acl
	BeforeEach(func() {
		db = database.NewTestDB()
	})

	AfterEach(func() {
		database.EmptyDatabase(db)
	})

	When("Read access to a list", func() {
		It("Grant access", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			err := doorKeeper.GrantListReadAccess("a", "b")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Revoke access", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantListReadAccess("a", "b")
			err := doorKeeper.RevokeListReadAccess("a", "b")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Has Read access after being granted", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantListReadAccess("a", "b")
			response, _ := doorKeeper.HasUserListReadAccess("a", "b")
			Expect(response).Should(Equal(true))
		})

		It("Does not have read access after being revoked", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantListReadAccess("a", "b")
			doorKeeper.RevokeListReadAccess("a", "b")
			response, _ := doorKeeper.HasUserListReadAccess("a", "b")
			Expect(response).Should(Equal(false))
		})
	})

	When("Write access to a list", func() {
		It("Grant access", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			err := doorKeeper.GrantListWriteAccess("a", "b")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Revoke access", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantListWriteAccess("a", "b")
			err := doorKeeper.RevokeListWriteAccess("a", "b")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Has Read access after being granted", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantListWriteAccess("a", "b")
			response, _ := doorKeeper.HasWriteAccess("a", "b")
			Expect(response).Should(Equal(true))
		})

		It("Does not have read access after being revoked", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantListWriteAccess("a", "b")
			doorKeeper.RevokeListWriteAccess("a", "b")
			response, _ := doorKeeper.HasWriteAccess("a", "b")
			Expect(response).Should(Equal(false))
		})
	})

	When("Sharing the list", func() {
		var allow bool
		BeforeEach(func() {
			allow = false
		})
		It("Make it public", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			err := doorKeeper.ShareListWithPublic("a")
			Expect(err).ShouldNot(HaveOccurred())

			allow, _ = doorKeeper.IsListAvailableToFriends("a")
			Expect(allow).To(Equal(false))
			allow, _ = doorKeeper.IsListPrivate("a")
			Expect(allow).To(Equal(false))
			allow, _ = doorKeeper.IsListPublic("a")
			Expect(allow).To(Equal(true))
			with, _ := doorKeeper.ListIsSharedWith("a")
			Expect(with).To(Equal(keys.SharedWithPublic))
		})

		It("Make it available to friends", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			err := doorKeeper.ShareListWithFriends("a")
			Expect(err).ShouldNot(HaveOccurred())
			allow, _ = doorKeeper.IsListPrivate("a")
			Expect(allow).To(Equal(false))
			allow, _ = doorKeeper.IsListPublic("a")
			Expect(allow).To(Equal(false))
			allow, _ = doorKeeper.IsListAvailableToFriends("a")
			Expect(allow).To(Equal(true))
			with, _ := doorKeeper.ListIsSharedWith("a")
			Expect(with).To(Equal(keys.SharedWithFriends))
		})

		It("Make it private", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			err := doorKeeper.ShareListWithPrivate("a")
			Expect(err).ShouldNot(HaveOccurred())

			allow, _ = doorKeeper.IsListPublic("a")
			Expect(allow).To(Equal(false))
			allow, _ = doorKeeper.IsListAvailableToFriends("a")
			Expect(allow).To(Equal(false))
			allow, _ = doorKeeper.IsListPrivate("a")
			Expect(allow).To(Equal(true))

			with, _ := doorKeeper.ListIsSharedWith("a")
			Expect(with).To(Equal(keys.NotShared))
		})
	})
})

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
		//database.EmptyDatabase(db)
	})

	When("Read access to a list", func() {
		It("Grant access", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			err := doorKeeper.GrantUserListReadAccess("a", "b")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Revoke access", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantUserListReadAccess("a", "b")
			err := doorKeeper.RevokeUserListReadAccess("a", "b")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Has Read access after being granted", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantUserListReadAccess("a", "b")
			response, _ := doorKeeper.HasUserListReadAccess("a", "b")
			Expect(response).Should(Equal(true))
		})

		It("Does not have read access after being revoked", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantUserListReadAccess("a", "b")
			doorKeeper.RevokeUserListReadAccess("a", "b")
			response, _ := doorKeeper.HasUserListReadAccess("a", "b")
			Expect(response).Should(Equal(false))
		})
	})

	When("Write access to a list", func() {
		It("Grant access", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			err := doorKeeper.GrantUserListWriteAccess("a", "b")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Revoke access", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantUserListWriteAccess("a", "b")
			err := doorKeeper.RevokeUserListWriteAccess("a", "b")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Has Read access after being granted", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantUserListWriteAccess("a", "b")
			response, _ := doorKeeper.HasUserListWriteAccess("a", "b")
			Expect(response).Should(Equal(true))
		})

		It("Does not have read access after being revoked", func() {
			doorKeeper = aclSqlite.NewAcl(db)
			doorKeeper.GrantUserListWriteAccess("a", "b")
			doorKeeper.RevokeUserListWriteAccess("a", "b")
			response, _ := doorKeeper.HasUserListWriteAccess("a", "b")
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
			err := doorKeeper.MakeListPrivate("a", "b")
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

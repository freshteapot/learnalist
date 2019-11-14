package sqlite_test

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	storage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/jmoiron/sqlx"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Acl", func() {
	var (
		db         *sqlx.DB
		repoistory user.UserFromIDP
	)
	BeforeEach(func() {
		db = database.NewTestDB()
	})

	AfterEach(func() {
		database.EmptyDatabase(db)
	})

	XIt("Testing with real queries", func() {
		repoistory = storage.NewUserFromIDP(db)
		userUUID, err := repoistory.Register("google", "fake@freshteapot.net", []byte(`{"name", "chris"}`))
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(userUUID)

		userUUID_2, err := repoistory.Lookup("google", "fake@freshteapot.net")
		fmt.Println(userUUID_2)
		fmt.Println(err)
	})
})

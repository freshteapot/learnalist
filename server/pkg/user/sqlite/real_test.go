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
		repository user.UserFromIDP
	)
	BeforeEach(func() {
		db = database.NewTestDB()
	})

	AfterEach(func() {
		database.EmptyDatabase(db)
	})

	XIt("Testing with real queries", func() {
		repository = storage.NewUserFromIDP(db)
		userUUID, err := repository.Register("google", "fake@freshteapot.net", []byte(`{"name", "chris"}`))
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(userUUID)

		userUUID_2, err := repository.Lookup("google", "fake@freshteapot.net")
		fmt.Println(userUUID_2)
		fmt.Println(err)
	})
})

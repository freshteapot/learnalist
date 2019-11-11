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

	It("Grant access", func() {
		repoistory = storage.NewUserFromIDP(db)
		userUUID, err := repoistory.Register("google", "chris@freshteapot.net", []byte(`{"name", "chris"}`))
		Expect(err).ShouldNot(HaveOccurred())
		fmt.Println(userUUID)

		userUUID_2, err := repoistory.Lookup("google", "chris@freshteapot.net")
		fmt.Println(userUUID_2)
		fmt.Println(err)
		userInfo, err := repoistory.GetByUserUUID(userUUID_2)
		fmt.Println(userInfo)
		fmt.Println(err)
	})
})

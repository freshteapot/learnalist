package api_test

import (
	"testing"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo/mocks"
	"github.com/freshteapot/learnalist-api/server/api/api"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Test Suite")
}

var dal *models.DAL
var m api.Manager

var _ = BeforeSuite(func() {
	db := database.NewTestDB()
	acl := aclStorage.NewAcl(db)
	userSession := userStorage.NewUserSession(db)
	userFromIDP := userStorage.NewUserFromIDP(db)
	oauthHandler := oauthStorage.NewOAuthReadWriter(db)
	dal = models.NewDAL(db, acl, userSession, userFromIDP, oauthHandler)
	hugoHelper := new(mocks.HugoSiteBuilder)

	m = api.Manager{
		Datastore:  dal,
		Acl:        acl,
		HugoHelper: hugoHelper,
	}
})

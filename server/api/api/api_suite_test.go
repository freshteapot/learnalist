package api_test

import (
	"testing"

	"github.com/freshteapot/learnalist-api/server/api/api"
	"github.com/freshteapot/learnalist-api/server/api/database"
	labelStorage "github.com/freshteapot/learnalist-api/server/api/label/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/mocks"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPackage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Api Test Suite")
}

var (
	dal *models.DAL
	m   api.Manager
	db  *sqlx.DB
)

var _ = BeforeSuite(func() {
	db = database.NewTestDB()
	acl := aclStorage.NewAcl(db)
	oauthHandler := oauthStorage.NewOAuthReadWriter(db)
	labels := labelStorage.NewLabel(db)
	aListStorage := &mocks.DatastoreAlists{}
	dal = models.NewDAL(acl, aListStorage, labels, oauthHandler)

	m = api.Manager{
		Datastore: dal,
		Acl:       acl,
	}
})

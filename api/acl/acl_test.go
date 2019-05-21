package acl

import (
	"testing"

	"github.com/freshteapot/learnalist-api/api/database"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

var db *sqlx.DB
var acl *Acl

type AclSuite struct {
	suite.Suite
}

func (suite *AclSuite) SetupSuite() {
	resetDatabase()
}

func (suite *AclSuite) SetupTest() {

}

func (suite *AclSuite) TearDownTest() {
	database.EmptyDatabase(db)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(AclSuite))
}

func resetDatabase() {
	db = database.NewTestDB()
	acl = NewAclFromModel(database.PathToTestSqliteDb)
}

package models

import (
	"fmt"
	"testing"
	"github.com/freshteapot/learnalist-api/api/acl"
	"github.com/stretchr/testify/suite"
)

var dal *DAL

type ModelSuite struct {
	suite.Suite
	UserUUID string
}

func (suite *ModelSuite) SetupSuite() {
	resetDatabase()
}

func (suite *ModelSuite) SetupTest() {
	suite.UserUUID = setupUserViaSQL()
}

func (suite *ModelSuite) TearDownTest() {
	tables := GetTables()
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		dal.Db.MustExec(query)
	}
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

func resetDatabase() {
	db, _ := NewTestDB()
	acl := acl.NewAclFromModel(PathToTestSqliteDb)
	dal = &DAL{
		Db:  db,
		Acl: acl,
	}
}

func setupUserViaSQL() string {
	setup := `
INSERT INTO user VALUES('7540fe5f-9847-5473-bdbd-2b20050da0c6','9046052444752556320','chris');
`
	dal.Db.MustExec(setup)
	return "7540fe5f-9847-5473-bdbd-2b20050da0c6"
}

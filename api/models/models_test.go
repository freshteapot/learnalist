package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

var dal *DAL

type ModelSuite struct {
	suite.Suite
}

func (suite *ModelSuite) SetupSuite() {
	resetDatabase()
}

func (suite *ModelSuite) SetupTest() {

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
	dal = &DAL{
		Db: db,
	}
}

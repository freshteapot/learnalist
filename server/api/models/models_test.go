package models_test

import (
	"testing"

	alistStorage "github.com/freshteapot/learnalist-api/server/api/alist/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/database"
	labelStorage "github.com/freshteapot/learnalist-api/server/api/label/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/models"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/suite"
)

var (
	dal *models.DAL
	db  *sqlx.DB
)

type ModelSuite struct {
	suite.Suite
	UserUUID string
}

func (suite *ModelSuite) SetupSuite() {
	logger, _ := test.NewNullLogger()
	db = database.NewTestDB()
	acl := aclStorage.NewAcl(db)
	labels := labelStorage.NewLabel(db)
	storageAlist := alistStorage.NewAlist(db, logger)

	dal = models.NewDAL(
		acl,
		storageAlist,
		labels,
	)
}

func (suite *ModelSuite) SetupTest() {
	suite.UserUUID = setupUserViaSQL()
}

func (suite *ModelSuite) TearDownTest() {
	database.EmptyDatabase(db)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

func setupUserViaSQL() string {
	setup := `
INSERT INTO user VALUES('7540fe5f-9847-5473-bdbd-2b20050da0c6','9046052444752556320','chris');
`
	db.MustExec(setup)
	return "7540fe5f-9847-5473-bdbd-2b20050da0c6"
}

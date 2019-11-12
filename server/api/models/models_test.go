package models

import (
	"testing"

	"github.com/freshteapot/learnalist-api/server/api/database"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/stretchr/testify/suite"
)

var dal *DAL

type ModelSuite struct {
	suite.Suite
	UserUUID string
}

func (suite *ModelSuite) SetupSuite() {
	db := database.NewTestDB()
	acl := aclStorage.NewAcl(db)
	userSession := userStorage.NewUserSession(db)
	userFromIDP := userStorage.NewUserFromIDP(db)
	oauthHandler := oauthStorage.NewOAuthReadWriter(db)
	dal = NewDAL(db, acl, userSession, userFromIDP, oauthHandler)
}

func (suite *ModelSuite) SetupTest() {
	suite.UserUUID = setupUserViaSQL()
}

func (suite *ModelSuite) TearDownTest() {
	database.EmptyDatabase(dal.Db)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

func setupUserViaSQL() string {
	setup := `
INSERT INTO user VALUES('7540fe5f-9847-5473-bdbd-2b20050da0c6','9046052444752556320','chris');
`
	dal.Db.MustExec(setup)
	return "7540fe5f-9847-5473-bdbd-2b20050da0c6"
}

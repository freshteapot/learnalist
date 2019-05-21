package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freshteapot/learnalist-api/api/acl"
	"github.com/freshteapot/learnalist-api/api/database"
	"github.com/freshteapot/learnalist-api/api/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

var dal *models.DAL
var env = Env{
	Port:         1234,
	DatabaseName: database.PathToTestSqliteDb,
}

type ApiSuite struct {
	suite.Suite
}

func (suite *ApiSuite) SetupSuite() {
	resetDatabase()
}

func (suite *ApiSuite) SetupTest() {

}

func (suite *ApiSuite) TearDownTest() {
	database.EmptyDatabase(dal.Db)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(ApiSuite))
}

func resetDatabase() {
	db := database.NewTestDB()
	acl := acl.NewAclFromModel(database.PathToTestSqliteDb)
	dal = models.NewDAL(db, acl)
	env.Datastore = dal
	env.Acl = *acl
}

func setupFakeEndpoint(method string, uri string, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, uri, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	return req, rec
}

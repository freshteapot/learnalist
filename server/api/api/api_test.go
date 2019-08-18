package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo/mocks"
	"github.com/freshteapot/learnalist-api/server/api/acl"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

var dal *models.DAL
var m Manager

type ApiSuite struct {
	suite.Suite
}

func (suite *ApiSuite) SetupSuite() {
	db := database.NewTestDB()
	acl := acl.NewAclFromModel(database.PathToTestSqliteDb)
	dal = models.NewDAL(db, acl)

	hugoHelper := new(mocks.HugoSiteBuilder)

	// setup expectations with a placeholder in the argument list

	m = Manager{
		Datastore:  dal,
		Acl:        *acl,
		HugoHelper: hugoHelper,
	}
}

func (suite *ApiSuite) SetupTest() {

}

func (suite *ApiSuite) TearDownTest() {
	database.EmptyDatabase(dal.Db)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(ApiSuite))
}

func setupFakeEndpoint(method string, uri string, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, uri, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	return req, rec
}

package api

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freshteapot/learnalist-api/api/models"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

var dal *models.DAL
var env = Env{
	Port:         9090,
	DatabaseName: "./test.db",
}

type ApiSuite struct {
	suite.Suite
}

func (suite *ApiSuite) SetupSuite() {
	// Init and set mysql cleanup engine
	fmt.Println("Setup")
}

func (suite *ApiSuite) SetupTest() {
	resetDatabase()
}

func (suite *ApiSuite) TearDownTest() {
	tables := models.GetTables()
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		dal.Db.MustExec(query)
	}
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(ApiSuite))
}

func resetDatabase() {
	db, err := models.NewTestDB()
	if err != nil {
		log.Panic(err)
	}
	dal = &models.DAL{
		Db: db,
	}
	env.Datastore = dal
}

func setupFakeEndpoint(method string, uri string, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, uri, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	return req, rec
}

package api

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/freshteapot/learnalist-api/api/models"
	"github.com/labstack/echo/v4"
)

var dal *models.DAL
var env = Env{
	Port:         9090,
	DatabaseName: "./test.db",
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

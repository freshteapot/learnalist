package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/gomega"
)

func emptyDatabase() {
	database.EmptyDatabase(db)
}

func createAList(userUUID, input string) (statusCode int, responseBytes []byte) {
	user := &uuid.User{
		Uuid: userUUID,
	}

	req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/alist", input)
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)

	err := m.V1SaveAlist(c)
	Expect(err).ShouldNot(HaveOccurred())
	return rec.Code, rec.Body.Bytes()
}

func setupAlistPostEndpoint(userUUID, input string) (echo.Context, *http.Request, *httptest.ResponseRecorder) {
	user := &uuid.User{
		Uuid: userUUID,
	}

	req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/alist", input)
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	return c, req, rec
}

func setupFakeEndpoint(method string, uri string, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, uri, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	return req, rec
}

func getValidUserRegisterInput(which string) string {
	if which == "b" {
		return `{"username":"iamuserb", "password":"test123"}`
	}

	return `{"username":"iamusera", "password":"test123"}`
}

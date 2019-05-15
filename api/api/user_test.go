package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/labstack/echo/v4"
)

// @todo this needs faking of the actual database commands.

func (suite *ApiSuite) TestPostRegisterEmptyBody() {
	expected := `{"message":"Bad input."}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if suite.NoError(env.PostRegister(c)) {
		suite.Equal(http.StatusBadRequest, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		suite.Equal(expected, response)
	}
}

func (suite *ApiSuite) TestPostRegisterNotValidJSON() {
	badInput := `{username:"chris", password:"test"}`
	expected := `{"message":"Bad input."}`

	req, rec := setupFakeEndpoint(http.MethodPost, "/register", badInput)
	e := echo.New()
	c := e.NewContext(req, rec)

	if suite.NoError(env.PostRegister(c)) {
		suite.Equal(http.StatusBadRequest, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		suite.Equal(expected, response)
	}
}

func (suite *ApiSuite) TestPostRegisterNotValidPayload() {
	badInput := `{"username":"", "password":""}`
	expected := `{"message":"Bad input."}`

	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/register", badInput)
	c := e.NewContext(req, rec)

	if suite.NoError(env.PostRegister(c)) {
		suite.Equal(http.StatusBadRequest, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		suite.Equal(expected, response)
	}
}

func (suite *ApiSuite) TestPostRegisterValidPayload() {
	input := `{"username":"chris", "password":"test"}`
	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/register", input)
	c := e.NewContext(req, rec)

	if suite.NoError(env.PostRegister(c)) {
		suite.Equal(http.StatusCreated, rec.Code)
		responseA := strings.TrimSpace(rec.Body.String())

		// Check we get the same userid
		req, rec := setupFakeEndpoint(http.MethodPost, "/register", input)
		c := e.NewContext(req, rec)

		if suite.NoError(env.PostRegister(c)) {
			suite.Equal(http.StatusOK, rec.Code)
			responseB := strings.TrimSpace(rec.Body.String())
			suite.Equal(responseA, responseB)
		}
	}
}

func (suite *ApiSuite) TestPostRegisterValidPayloadThenFake() {
	input := `{"username":"chris", "password":"test"}`
	fake := `{"username":"chris", "password":"test123"}`
	expectedFakeResponse := fmt.Sprintf(`{"message":"%s"}`, i18n.UserInsertUsernameExists)
	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/register", input)
	c := e.NewContext(req, rec)

	if suite.NoError(env.PostRegister(c)) {
		suite.Equal(http.StatusCreated, rec.Code)

		// Check we get the same userid
		req, rec := setupFakeEndpoint(http.MethodPost, "/register", fake)
		c := e.NewContext(req, rec)

		if suite.NoError(env.PostRegister(c)) {
			suite.Equal(http.StatusBadRequest, rec.Code)
			response := strings.TrimSpace(rec.Body.String())
			suite.Equal(response, expectedFakeResponse)
		}
	}
}

func (suite *ApiSuite) TestPostRegisterRepeat() {
	input := `{"username":"chris", "password":"test"}`
	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/register", input)
	c := e.NewContext(req, rec)

	env.PostRegister(c)
	suite.Equal(http.StatusCreated, rec.Code)

	req, rec = setupFakeEndpoint(http.MethodPost, "/register", input)
	c = e.NewContext(req, rec)
	env.PostRegister(c)
	suite.Equal(http.StatusOK, rec.Code)
}

package api

import (
	"encoding/json"
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
	var statusCode int
	input := `{"username":"chris", "password":"test"}`
	_, statusCode = suite.createNewUserWithSuccess(input)
	suite.Equal(http.StatusCreated, statusCode)

	_, statusCode = suite.createNewUserWithSuccess(input)
	suite.Equal(http.StatusOK, statusCode)
}

func (suite *ApiSuite) createNewUserWithSuccess(input string) (uuid string, httpStatusCode int) {

	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/register", input)
	c := e.NewContext(req, rec)
	suite.NoError(env.PostRegister(c))
	suite.Contains([]int{http.StatusOK, http.StatusCreated}, rec.Code)

	var raw map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &raw)
	user_uuid := raw["uuid"].(string)
	return user_uuid, rec.Code
}

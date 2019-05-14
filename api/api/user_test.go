package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// @todo this needs faking of the actual database commands.

func init() {
	resetDatabase()
}

func TestPostRegisterEmptyBody(t *testing.T) {
	resetDatabase()
	expected := `{"message":"Bad input."}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if assert.NoError(t, env.PostRegister(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		assert.Equal(t, expected, response)
	}
}

func TestPostRegisterNotValidJSON(t *testing.T) {
	resetDatabase()
	badInput := `{username:"chris", password:"test"}`
	expected := `{"message":"Bad input."}`

	req, rec := setupFakeEndpoint(http.MethodPost, "/register", badInput)
	e := echo.New()
	c := e.NewContext(req, rec)

	if assert.NoError(t, env.PostRegister(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		assert.Equal(t, expected, response)
	}
}

func TestPostRegisterNotValidPayload(t *testing.T) {
	resetDatabase()
	badInput := `{"username":"", "password":""}`
	expected := `{"message":"Bad input."}`

	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/register", badInput)
	c := e.NewContext(req, rec)

	if assert.NoError(t, env.PostRegister(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		assert.Equal(t, expected, response)
	}
}

func TestPostRegisterValidPayload(t *testing.T) {
	resetDatabase()
	input := `{"username":"chris", "password":"test"}`
	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/register", input)
	c := e.NewContext(req, rec)

	if assert.NoError(t, env.PostRegister(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		responseA := strings.TrimSpace(rec.Body.String())

		// Check we get the same userid
		req, rec := setupFakeEndpoint(http.MethodPost, "/register", input)
		c := e.NewContext(req, rec)

		if assert.NoError(t, env.PostRegister(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			responseB := strings.TrimSpace(rec.Body.String())
			assert.Equal(t, responseA, responseB)
		}
	}
}

func TestPostRegisterValidPayloadThenFake(t *testing.T) {
	resetDatabase()
	input := `{"username":"chris", "password":"test"}`
	fake := `{"username":"chris", "password":"test123"}`
	expectedFakeResponse := `{"message":"Failed to save."}`
	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/register", input)
	c := e.NewContext(req, rec)

	if assert.NoError(t, env.PostRegister(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Check we get the same userid
		req, rec := setupFakeEndpoint(http.MethodPost, "/register", fake)
		c := e.NewContext(req, rec)

		if assert.NoError(t, env.PostRegister(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			response := strings.TrimSpace(rec.Body.String())
			assert.Equal(t, response, expectedFakeResponse)
		}
	}
}

func TestPostRegisterRepeat(t *testing.T) {
	resetDatabase()
	input := `{"username":"chris", "password":"test"}`
	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/register", input)
	c := e.NewContext(req, rec)

	env.PostRegister(c)
	assert.Equal(t, http.StatusCreated, rec.Code)

	req, rec = setupFakeEndpoint(http.MethodPost, "/register", input)
	c = e.NewContext(req, rec)
	env.PostRegister(c)
	assert.Equal(t, http.StatusOK, rec.Code)
}

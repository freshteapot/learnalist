package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func init() {
	resetDatabase()
}

func TestPostLabel(t *testing.T) {
	resetDatabase()
	inputA := `{"label": "car"}`
	inputB := `{"label": "boat"}`

	req, rec := setupFakeEndpoint(http.MethodPost, "/labels", inputA)
	e := echo.New()
	c := e.NewContext(req, rec)

	user := uuid.NewUser()
	c.Set("loggedInUser", user)
	assert.NoError(t, env.PostUserLabel(c))
	assert.Equal(t, http.StatusCreated, rec.Code)

	req, rec = setupFakeEndpoint(http.MethodPost, "/labels", inputA)
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.PostUserLabel(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	req, rec = setupFakeEndpoint(http.MethodPost, "/labels", inputB)
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.PostUserLabel(c))
	assert.Equal(t, http.StatusCreated, rec.Code)
}

func TestGetUsersLabels(t *testing.T) {
	resetDatabase()
	var req *http.Request
	var rec *httptest.ResponseRecorder
	var c echo.Context
	e := echo.New()
	user := uuid.NewUser()

	req, rec = setupFakeEndpoint(http.MethodGet, "/labels/by/me", "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetUserLabels(c))
	response := strings.TrimSpace(rec.Body.String())
	// Check it is an empty array
	assert.Equal(t, "[]", response)

	// Post a label
	input := []string{
		`{"label": "car"}`,
		`{"label": "boat"}`,
		`{"label": "car"}`,
	}
	for _, item := range input {
		req, rec = setupFakeEndpoint(http.MethodPost, "/labels", item)
		c = e.NewContext(req, rec)
		c.Set("loggedInUser", user)
		assert.NoError(t, env.PostUserLabel(c))
	}

	req, rec = setupFakeEndpoint(http.MethodGet, "/labels/by/me", "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetUserLabels(c))
	response = strings.TrimSpace(rec.Body.String())
	assert.Equal(t, `["boat","car"]`, response)
}

func TestDeleteUsersLabels(t *testing.T) {
	resetDatabase()
	var req *http.Request
	var rec *httptest.ResponseRecorder
	var c echo.Context
	e := echo.New()
	user := uuid.NewUser()

	req, rec = setupFakeEndpoint(http.MethodDelete, "/labels/car", "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.RemoveUserLabel(c))
	response := strings.TrimSpace(rec.Body.String())
	fmt.Println(response)
}

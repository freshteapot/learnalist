package api

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
)

func (suite *ApiSuite) TestPostLabel() {
	inputA := `{"label": "car"}`
	inputB := `{"label": "boat"}`

	req, rec := setupFakeEndpoint(http.MethodPost, "/labels", inputA)
	e := echo.New()
	c := e.NewContext(req, rec)

	user := uuid.NewUser()
	c.Set("loggedInUser", user)
	suite.NoError(env.PostUserLabel(c))
	suite.Equal(http.StatusCreated, rec.Code)

	req, rec = setupFakeEndpoint(http.MethodPost, "/labels", inputA)
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.PostUserLabel(c))
	suite.Equal(http.StatusOK, rec.Code)

	req, rec = setupFakeEndpoint(http.MethodPost, "/labels", inputB)
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.PostUserLabel(c))
	suite.Equal(http.StatusCreated, rec.Code)
}

func (suite *ApiSuite) TestGetUsersLabels() {
	var req *http.Request
	var rec *httptest.ResponseRecorder
	var c echo.Context
	e := echo.New()
	user := uuid.NewUser()

	req, rec = setupFakeEndpoint(http.MethodGet, "/labels/by/me", "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.GetUserLabels(c))
	response := strings.TrimSpace(rec.Body.String())
	// Check it is an empty array
	suite.Equal("[]", response)

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
		suite.NoError(env.PostUserLabel(c))
	}

	req, rec = setupFakeEndpoint(http.MethodGet, "/labels/by/me", "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.GetUserLabels(c))
	response = strings.TrimSpace(rec.Body.String())
	suite.Equal(`["boat","car"]`, response)
}

func (suite *ApiSuite) TestDeleteUsersLabels() {
	var req *http.Request
	var rec *httptest.ResponseRecorder
	var c echo.Context
	e := echo.New()
	user := uuid.NewUser()

	req, rec = setupFakeEndpoint(http.MethodDelete, "/labels/car", "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.RemoveUserLabel(c))
	response := strings.TrimSpace(rec.Body.String())
	suite.Equal(`{"message":"Label car was removed."}`, response)
}

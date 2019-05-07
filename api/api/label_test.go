package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/freshteapot/learnalist-api/api/models"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func init() {
	resetDatabase()
}

func TestRemoveLabel(t *testing.T) {
	resetDatabase()

	input := `{"label": "car"}`

	req, rec := setupFakeEndpoint(http.MethodPost, "/labels", input)
	e := echo.New()
	c := e.NewContext(req, rec)

	user := uuid.NewUser()
	c.Set("loggedInUser", user)
	assert.NoError(t, env.PostLabel(c))
	assert.Equal(t, http.StatusCreated, rec.Code)
	responseA := strings.TrimSpace(rec.Body.String())
	labelA := models.Label{}
	json.Unmarshal([]byte(responseA), &labelA)

	// See it is there
	req, rec = setupFakeEndpoint(http.MethodGet, "/labels/by/me", "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetLabelsByUser(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	responseB := strings.TrimSpace(rec.Body.String())
	labelsB := []models.Label{}
	json.Unmarshal([]byte(responseB), &labelsB)

	assert.Equal(t, labelA, labelsB[0])

	// Remove the label
	uri := "/labels/" + labelA.Uuid
	req, rec = setupFakeEndpoint(http.MethodDelete, uri, "")
	c = e.NewContext(req, rec)

	c.Set("loggedInUser", user)
	assert.NoError(t, env.RemoveLabel(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	responseC := strings.TrimSpace(rec.Body.String())
	responseStruct := &HttpResponseMessage{}
	json.Unmarshal([]byte(responseC), &responseStruct)
	assert.Equal(t, fmt.Sprintf("Label %s was removed.", labelA.Uuid), responseStruct.Message)
}

func TestPostLabelBadData(t *testing.T) {
	resetDatabase()
	input := `"car"`

	req, rec := setupFakeEndpoint(http.MethodPost, "/labels", input)
	e := echo.New()
	c := e.NewContext(req, rec)

	user := uuid.NewUser()
	c.Set("loggedInUser", user)
	if assert.NoError(t, env.PostLabel(c)) {
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	}
}

func TestPostLabel(t *testing.T) {
	var response1 string
	var response2 string
	resetDatabase()
	input := `{"label": "car"}`

	req, rec := setupFakeEndpoint(http.MethodPost, "/labels", input)
	e := echo.New()
	c := e.NewContext(req, rec)

	user := uuid.NewUser()
	c.Set("loggedInUser", user)
	if assert.NoError(t, env.PostLabel(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		response1 = strings.TrimSpace(rec.Body.String())

		var raw map[string]interface{}
		json.Unmarshal([]byte(response1), &raw)
		assert.Equal(t, "car", raw["label"].(string))
	}

	// Check duplicate
	req, rec = setupFakeEndpoint(http.MethodPost, "/labels", input)
	e = echo.New()
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)

	assert.NoError(t, env.PostLabel(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	response2 = strings.TrimSpace(rec.Body.String())
	assert.Equal(t, response1, response2)
}

func TestGetLabelsByUserWithNoLabels(t *testing.T) {
	resetDatabase()
	expectedEmpty := `[]`

	req, rec := setupFakeEndpoint(http.MethodGet, "/labels/by/me", "")
	e := echo.New()
	c := e.NewContext(req, rec)

	user := uuid.NewUser()
	c.Set("loggedInUser", user)

	if assert.NoError(t, env.GetLabelsByUser(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		assert.Equal(t, expectedEmpty, response)
	}
}

func TestGetLabelsByUserWithLabels(t *testing.T) {
	resetDatabase()
	expectedEmpty := `[]`

	req, rec := setupFakeEndpoint(http.MethodPost, "/labels/by/me", expectedEmpty)
	e := echo.New()
	c := e.NewContext(req, rec)

	user := uuid.NewUser()
	c.Set("loggedInUser", user)

	if assert.NoError(t, env.GetLabelsByUser(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		assert.Equal(t, expectedEmpty, response)
	}
}

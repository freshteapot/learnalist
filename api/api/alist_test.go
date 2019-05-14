package api

import (
	"encoding/json"
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

func TestAlistApi(t *testing.T) {
	resetDatabase()
	// Post a list
	lists := []string{`
{
	"data": ["car"],
	"info": {
		"title": "Days of the Week",
		"type": "v1"
	}
}
`,
		`
{
	"data": [],
	"info": {
		"title": "Days of the Week 2",
		"type": "v1"
	}
}
`,
		`{
	"data": ["car"],
	"info": {
		"title": "Days of the Week",
		"type": "v1",
		"labels": [
			"car",
			"water"
		]
	}
}`,
		`{
	"data": [{"from":"car", "to": "bil"}],
	"info": {
		"title": "Days of the Week",
		"type": "v2",
			"labels": [
			"water"
		]
	}
}`,
	}
	uuids := []string{}
	type uuidOnly struct {
		Uuid string `json:"uuid"`
	}
	type listOfUuidsOnly []uuidOnly

	var listUuid uuidOnly
	var req *http.Request
	var rec *httptest.ResponseRecorder
	var e *echo.Echo
	var c echo.Context
	var uri string
	var response string
	var listOfUuids listOfUuidsOnly

	user := uuid.NewUser()
	uri = "/alist"
	for _, item := range lists {
		req, rec = setupFakeEndpoint(http.MethodPost, uri, item)
		e = echo.New()
		c = e.NewContext(req, rec)
		c.Set("loggedInUser", user)

		assert.NoError(t, env.SaveAlist(c))
		assert.Equal(t, http.StatusCreated, rec.Code)
		response := strings.TrimSpace(rec.Body.String())

		json.Unmarshal([]byte(response), &listUuid)
		uuids = append(uuids, listUuid.Uuid)
	}
	fmt.Println(uuids)
	// Check a valid uuid
	uri = "/alist/" + uuids[0]
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetListByUUID(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	// Check an empty uuid
	uri = "/alist/"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetListByUUID(c))
	response = strings.TrimSpace(rec.Body.String())
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, `{"message":"The uuid is missing."}`, response)

	uri = "/alist/fake123"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetListByUUID(c))
	response = strings.TrimSpace(rec.Body.String())
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.True(t, strings.Contains(response, "Failed to find alist with uuid:"))

	// Get my lists
	uri = "/alist/by/me"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetListsByMe(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	assert.Equal(t, 4, len(listOfUuids))

	// Get my lists filter by labels
	uri = "/alist/by/me?labels=water"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetListsByMe(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	assert.Equal(t, 2, len(listOfUuids))

	uri = "/alist/by/me?labels="
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetListsByMe(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	assert.Equal(t, 0, len(listOfUuids))

	uri = "/alist/by/me?labels=car,water"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetListsByMe(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	assert.Equal(t, 2, len(listOfUuids))

	uri = "/alist/by/me?labels=card"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.GetListsByMe(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	assert.Equal(t, 0, len(listOfUuids))

	// Update a list
	putListData := `
{
	"data": [{"from":"car", "to": "bil"}],
	"info": {
	"title": "Updated",
	"type": "v2",
		"labels": [
		"water"
	]
	}
}
`
	uri = "/alist/" + uuids[0]
	req, rec = setupFakeEndpoint(http.MethodPut, uri, putListData)
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.SaveAlist(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	// Check bad data
	uri = "/alist/" + uuids[0]
	req, rec = setupFakeEndpoint(http.MethodPut, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.SaveAlist(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response = strings.TrimSpace(rec.Body.String())
	assert.Equal(t, `{"message":"Your Json has a problem. Failed to parse list."}`, response)
	// Check unsupported method
	uri = "/alist/" + uuids[0]
	req, rec = setupFakeEndpoint(http.MethodDelete, uri, putListData)
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.SaveAlist(c))
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	response = strings.TrimSpace(rec.Body.String())
	assert.Equal(t, `{"message":"This method is not supported."}`, response)

	// RemoveAlist
	uri = "/alist/" + uuids[0]
	req, rec = setupFakeEndpoint(http.MethodDelete, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	assert.NoError(t, env.RemoveAlist(c))
	assert.Equal(t, http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())
	assert.Equal(t, fmt.Sprintf(`{"message":"List %s was removed."}`, uuids[0]), response)
}

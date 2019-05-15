package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
)

func (suite *ApiSuite) TestAlistApi() {
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

		suite.NoError(env.SaveAlist(c))
		suite.Equal(http.StatusCreated, rec.Code)
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
	suite.NoError(env.GetListByUUID(c))
	suite.Equal(http.StatusOK, rec.Code)
	// Check an empty uuid
	uri = "/alist/"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.GetListByUUID(c))
	response = strings.TrimSpace(rec.Body.String())
	suite.Equal(http.StatusNotFound, rec.Code)
	suite.Equal(`{"message":"The uuid is missing."}`, response)

	uri = "/alist/fake123"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.GetListByUUID(c))
	response = strings.TrimSpace(rec.Body.String())
	suite.Equal(http.StatusNotFound, rec.Code)
	suite.True(strings.Contains(response, "Failed to find alist with uuid:"))

	// Get my lists
	uri = "/alist/by/me"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.GetListsByMe(c))
	suite.Equal(http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	suite.Equal(4, len(listOfUuids))

	// Get my lists filter by labels
	uri = "/alist/by/me?labels=water"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.GetListsByMe(c))
	suite.Equal(http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	suite.Equal(2, len(listOfUuids))

	uri = "/alist/by/me?labels="
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.GetListsByMe(c))
	suite.Equal(http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	suite.Equal(0, len(listOfUuids))

	uri = "/alist/by/me?labels=car,water"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.GetListsByMe(c))
	suite.Equal(http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	suite.Equal(2, len(listOfUuids))

	uri = "/alist/by/me?labels=card"
	req, rec = setupFakeEndpoint(http.MethodGet, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.GetListsByMe(c))
	suite.Equal(http.StatusOK, rec.Code)
	response = strings.TrimSpace(rec.Body.String())

	json.Unmarshal([]byte(response), &listOfUuids)
	suite.Equal(0, len(listOfUuids))

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
	suite.NoError(env.SaveAlist(c))
	suite.Equal(http.StatusOK, rec.Code)
	// Check bad data
	uri = "/alist/" + uuids[0]
	req, rec = setupFakeEndpoint(http.MethodPut, uri, "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.SaveAlist(c))
	suite.Equal(http.StatusBadRequest, rec.Code)
	response = strings.TrimSpace(rec.Body.String())
	suite.Equal(`{"message":"Your Json has a problem. Failed to parse list."}`, response)
	// Check unsupported method
	uri = "/alist/" + uuids[0]
	req, rec = setupFakeEndpoint(http.MethodDelete, uri, putListData)
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.SaveAlist(c))
	suite.Equal(http.StatusBadRequest, rec.Code)
	response = strings.TrimSpace(rec.Body.String())
	suite.Equal(`{"message":"This method is not supported."}`, response)

	// RemoveAlist
	var raw map[string]interface{}
	statusCode, responseBytes := suite.removeAlist(user.Uuid, uuids[0])
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(fmt.Sprintf(i18n.ApiDeleteAlistSuccess, uuids[0]), raw["message"].(string))
}

// Linked to https://github.com/freshteapot/learnalist-api/issues/12.
func (suite *ApiSuite) TestOnlyOwnerOfTheListCanAlterIt() {
	inputUserA := `{"username":"iamusera", "password":"test"}`
	inputUserB := `{"username":"iamuserb", "password":"test"}`
	inputAlist := `
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
	user_uuidA, _ := suite.createNewUserWithSuccess(inputUserA)
	user_uuidB, _ := suite.createNewUserWithSuccess(inputUserB)
	suite.NotEmpty(user_uuidA, "It hopefully is a real user.")
	suite.NotEmpty(user_uuidB, "It hopefully is a real user.")
	var raw map[string]interface{}
	var statusCode int
	var responseBytes []byte

	responseBytes = suite.createAList(user_uuidA, inputAlist)
	json.Unmarshal(responseBytes, &raw)
	alist_uuid := raw["uuid"].(string)
	suite.NotEmpty(alist_uuid, "It hopefully is a real list.")

	responseBytes, _ = json.Marshal(raw)
	statusCode, responseBytes = suite.updateAlist(user_uuidA, alist_uuid, string(responseBytes))
	suite.Equal(http.StatusOK, statusCode)
	// User B should not be able to update list from user A.
	statusCode, responseBytes = suite.updateAlist(user_uuidB, alist_uuid, string(responseBytes))
	suite.Equal(http.StatusForbidden, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.InputSaveAlistOperationOwnerOnly, raw["message"].(string))

	// User B trying to delete User A list
	statusCode, responseBytes = suite.removeAlist(user_uuidB, alist_uuid)
	suite.Equal(http.StatusForbidden, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.InputDeleteAlistOperationOwnerOnly, raw["message"].(string))
}

func (suite *ApiSuite) TestDeleteAlistNotFound() {
	var raw map[string]interface{}
	inputUserA := `{"username":"iamusera", "password":"test"}`
	user_uuidA, _ := suite.createNewUserWithSuccess(inputUserA)
	alist_uuid := "fake"
	statusCode, responseBytes := suite.removeAlist(user_uuidA, alist_uuid)
	suite.Equal(http.StatusNotFound, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.SuccessAlistNotFound, raw["message"].(string))

}

func (suite *ApiSuite) updateAlist(user_uuid, alist_uuid string, input string) (statusCode int, body []byte) {
	method := http.MethodPut
	uri := fmt.Sprintf("/alist/%s", alist_uuid)
	user := &uuid.User{
		Uuid: user_uuid,
	}

	req, rec := setupFakeEndpoint(method, uri, input)
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	suite.NoError(env.SaveAlist(c))
	return rec.Code, rec.Body.Bytes()
}

func (suite *ApiSuite) createAList(user_uuid, input string) []byte {
	user := &uuid.User{
		Uuid: user_uuid,
	}

	req, rec := setupFakeEndpoint(http.MethodPost, "/alist", input)
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)

	suite.NoError(env.SaveAlist(c))
	suite.Equal(http.StatusCreated, rec.Code)
	return rec.Body.Bytes()
}

func (suite *ApiSuite) removeAlist(user_uuid string, alist_uuid string) (statusCode int, responseBytes []byte) {
	method := http.MethodDelete
	uri := fmt.Sprintf("/alist/%s", alist_uuid)

	user := &uuid.User{
		Uuid: user_uuid,
	}
	e := echo.New()
	req, rec := setupFakeEndpoint(method, uri, "")
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	suite.NoError(env.RemoveAlist(c))
	return rec.Code, rec.Body.Bytes()
}

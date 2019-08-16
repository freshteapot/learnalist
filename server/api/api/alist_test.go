package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

func (suite *ApiSuite) TestAlistV3Api() {
	var raw map[string]interface{}
	var statusCode int
	var responseBytes []byte
	input := `
{
  "info": {
      "title": "Getting my row on.",
      "type": "v3"
  },
  "data": [{
    "when": "2019-05-06",
    "overall": {
      "time": "7:15.9",
      "distance": 2000,
      "spm": 28,
      "p500": "1:48.9"
    },
    "splits": [
      {
        "time": "1:46.4",
        "distance": 500,
        "spm": 29,
        "p500": "1:58.0"
      }
    ]
  }]
}
`
	var aList alist.Alist
	inputUserA := getValidUserRegisterInput("a")
	userUUID, _ := suite.createNewUserWithSuccess(inputUserA)
	statusCode, responseBytes = suite.createAList(userUUID, input)
	suite.Equal(http.StatusCreated, statusCode)
	json.Unmarshal(responseBytes, &raw)
	json.Unmarshal(responseBytes, &aList)

	// Make sure the labels stay.
	aList.Info.Labels = nil
	updatedBytes, _ := aList.MarshalJSON()
	statusCode, _ = suite.updateAlist(userUUID, aList.Uuid, string(updatedBytes))

	suite.Equal(http.StatusOK, statusCode)
}

func (suite *ApiSuite) TestAlistApi() {
	var raw map[string]interface{}
	var statusCode int
	var responseBytes []byte
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
	var listOfUuids listOfUuidsOnly
	var listUuid uuidOnly

	inputUserA := getValidUserRegisterInput("a")
	userUUID, _ := suite.createNewUserWithSuccess(inputUserA)

	for _, item := range lists {
		statusCode, responseBytes = suite.createAList(userUUID, item)
		suite.Equal(http.StatusCreated, statusCode)
		json.Unmarshal(responseBytes, &listUuid)
		uuids = append(uuids, listUuid.Uuid)
	}

	// Check a valid uuid
	statusCode, responseBytes = suite.getList(userUUID, uuids[0])
	suite.Equal(http.StatusOK, statusCode)

	// Check an empty uuid
	statusCode, responseBytes = suite.getList(userUUID, "")
	suite.Equal(http.StatusNotFound, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.InputMissingListUuid, raw["message"].(string))
	raw = nil

	statusCode, responseBytes = suite.getList(userUUID, "fake")
	suite.Equal(http.StatusNotFound, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(fmt.Sprintf(i18n.ApiAlistNotFound, "fake"), raw["message"].(string))
	raw = nil
	// Get my lists
	statusCode, responseBytes = suite.getListsByMe(userUUID, "", "")
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &listOfUuids)
	suite.Equal(4, len(listOfUuids))

	// Get my lists filter by labels
	statusCode, responseBytes = suite.getListsByMe(userUUID, "water", "")
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &listOfUuids)
	suite.Equal(2, len(listOfUuids))

	// Check filter via listType works.
	statusCode, responseBytes = suite.getListsByMe(userUUID, "", alist.SimpleList)
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &listOfUuids)
	suite.Equal(3, len(listOfUuids))

	statusCode, responseBytes = suite.getListsByMe(userUUID, "", alist.FromToList)
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &listOfUuids)
	suite.Equal(1, len(listOfUuids))

	statusCode, responseBytes = suite.getListsByMe(userUUID, "", "")
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &listOfUuids)
	suite.Equal(4, len(listOfUuids))

	statusCode, responseBytes = suite.getListsByMe(userUUID, "car,water", "")
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &listOfUuids)
	suite.Equal(2, len(listOfUuids))

	// Check filter via labels and listType works.
	statusCode, responseBytes = suite.getListsByMe(userUUID, "car,water", "v2")
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &listOfUuids)
	suite.Equal(1, len(listOfUuids))

	statusCode, responseBytes = suite.getListsByMe(userUUID, "card", "")
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &listOfUuids)
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

	statusCode, responseBytes = suite.updateAlist(userUUID, uuids[0], putListData)
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(uuids[0], raw["uuid"].(string))
	raw = nil

	// Check bad data
	statusCode, responseBytes = suite.updateAlist(userUUID, uuids[0], "")
	suite.Equal(http.StatusBadRequest, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal("Your Json has a problem. Failed to parse list.", raw["message"].(string))
	raw = nil

	// RemoveAlist
	statusCode, responseBytes = suite.removeAlist(userUUID, uuids[0])
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(fmt.Sprintf(i18n.ApiDeleteAlistSuccess, uuids[0]), raw["message"].(string))
	raw = nil

}

func (suite *ApiSuite) TestMethodNotSupportedForSavingList() {
	user := &uuid.User{
		Uuid: "fake",
	}
	// Check unsupported method
	uri := "/alist/doesntmatter"
	req, rec := setupFakeEndpoint(http.MethodDelete, uri, "Doesnt matter")
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	suite.NoError(m.V1SaveAlist(c))
	suite.Equal(http.StatusBadRequest, rec.Code)
	response := strings.TrimSpace(rec.Body.String())
	suite.Equal(`{"message":"This method is not supported."}`, response)
}

// Linked to https://github.com/freshteapot/learnalist-api/issues/12.
func (suite *ApiSuite) TestOnlyOwnerOfTheListCanAlterIt() {
	inputUserA := getValidUserRegisterInput("a")
	inputUserB := getValidUserRegisterInput("b")
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

	statusCode, responseBytes = suite.createAList(user_uuidA, inputAlist)
	suite.Equal(http.StatusCreated, statusCode)
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
	raw = nil
	// User B trying to delete User A list
	statusCode, responseBytes = suite.removeAlist(user_uuidB, alist_uuid)
	suite.Equal(http.StatusForbidden, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.InputDeleteAlistOperationOwnerOnly, raw["message"].(string))
	raw = nil
}

func (suite *ApiSuite) TestDeleteAlistNotFound() {
	var raw map[string]interface{}
	inputUser := getValidUserRegisterInput("a")
	userUUID, _ := suite.createNewUserWithSuccess(inputUser)
	alistUUID := "fake"
	statusCode, responseBytes := suite.removeAlist(userUUID, alistUUID)
	suite.Equal(http.StatusNotFound, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.SuccessAlistNotFound, raw["message"].(string))
	raw = nil
}

func (suite *ApiSuite) updateAlist(userUUID, alistUUID string, input string) (statusCode int, body []byte) {
	method := http.MethodPut
	uri := fmt.Sprintf("/v1/alist/%s", alistUUID)
	user := &uuid.User{
		Uuid: userUUID,
	}

	req, rec := setupFakeEndpoint(method, uri, input)
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	suite.NoError(m.V1SaveAlist(c))
	return rec.Code, rec.Body.Bytes()
}

func (suite *ApiSuite) createAList(userUUID, input string) (statusCode int, responseBytes []byte) {
	user := &uuid.User{
		Uuid: userUUID,
	}

	req, rec := setupFakeEndpoint(http.MethodPost, "/v1/alist", input)
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)

	suite.NoError(m.V1SaveAlist(c))
	return rec.Code, rec.Body.Bytes()
}

func (suite *ApiSuite) removeAlist(userUUID string, alistUUID string) (statusCode int, responseBytes []byte) {
	method := http.MethodDelete
	uri := fmt.Sprintf("/v1/alist/%s", alistUUID)

	user := &uuid.User{
		Uuid: userUUID,
	}
	e := echo.New()
	req, rec := setupFakeEndpoint(method, uri, "")
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	suite.NoError(m.V1RemoveAlist(c))
	return rec.Code, rec.Body.Bytes()
}

func (suite *ApiSuite) getListsByMe(userUUID, labels string, listType string) (statusCode int, responseBytes []byte) {
	method := http.MethodGet
	uri := "/v1/alist/by/me"
	if labels != "" {
		uri = fmt.Sprintf("/v1/alist/by/me?labels=%s", labels)
	}
	if listType != "" {
		uri = fmt.Sprintf("/v1/alist/by/me?list_type=%s", listType)
	}
	if listType != "" && labels != "" {
		uri = fmt.Sprintf("/v1/alist/by/me?labels=%s&list_type=%s", labels, listType)
	}

	user := &uuid.User{
		Uuid: userUUID,
	}
	e := echo.New()
	req, rec := setupFakeEndpoint(method, uri, "")
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	suite.NoError(m.V1GetListsByMe(c))
	return rec.Code, rec.Body.Bytes()
}

func (suite *ApiSuite) getList(userUUID, alistUUID string) (statusCode int, responseBytes []byte) {
	method := http.MethodGet
	uri := fmt.Sprintf("/v1/alist/%s", alistUUID)
	user := &uuid.User{
		Uuid: userUUID,
	}
	e := echo.New()
	req, rec := setupFakeEndpoint(method, uri, "")
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	suite.NoError(m.V1GetListByUUID(c))
	return rec.Code, rec.Body.Bytes()
}

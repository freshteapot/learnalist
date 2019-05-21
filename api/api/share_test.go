package api

import (
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
)

func (suite *ApiSuite) TestV1ShareAlistBadInput() {
	inputUserA := `
{
  "username":"iamusera",
  "password":"test"
}
`
	userAUUID, _ := suite.createNewUserWithSuccess(inputUserA)
	var raw map[string]interface{}
	var statusCode int
	var responseBytes []byte

	statusCode, responseBytes = suite.shareAlist(userAUUID, "")
	suite.Equal(http.StatusBadRequest, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.PostShareListJSONFailure, raw["message"].(string))

	// Lookup list that doesnt exist
	inputGrant := &HttpShareListWithUserInput{
		UserUUID:  "fakeUser",
		AlistUUID: "fakeList",
		Action:    "grant",
	}
	a, _ := json.Marshal(inputGrant)
	statusCode, responseBytes = suite.shareAlist(userAUUID, string(a))
	suite.Equal(http.StatusNotFound, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.SuccessAlistNotFound, raw["message"].(string))

	// Check, list exists but user does not.
	inputAlist := `
{
  "data": [
    "monday",
    "tuesday",
    "wednesday",
    "thursday",
    "friday",
    "saturday",
    "sunday"
  ],
  "info": {
    "title": "Days of the Week",
    "type": "v1"
  }
}
`
	statusCode, responseBytes = suite.createAList(userAUUID, inputAlist)
	suite.Equal(http.StatusCreated, statusCode)
	json.Unmarshal(responseBytes, &raw)
	alistUUID := raw["uuid"].(string)
	inputGrant.AlistUUID = alistUUID
	a, _ = json.Marshal(inputGrant)
	statusCode, responseBytes = suite.shareAlist(userAUUID, string(a))
	suite.Equal(http.StatusNotFound, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.SuccessUserNotFound, raw["message"].(string))
}

func (suite *ApiSuite) TestV1ShareAlist() {
	inputUserA := `
{
  "username":"iamusera",
  "password":"test"
}
`
	inputUserB := `
{
"username":"iamuserb",
"password":"test"
}
`
	inputAlist := `
{
  "data": [
    "monday",
    "tuesday",
    "wednesday",
    "thursday",
    "friday",
    "saturday",
    "sunday"
  ],
  "info": {
    "title": "Days of the Week",
    "type": "v1"
  }
}
`
	userAUUID, _ := suite.createNewUserWithSuccess(inputUserA)
	userBUUID, _ := suite.createNewUserWithSuccess(inputUserB)
	var raw map[string]interface{}
	var statusCode int
	var responseBytes []byte

	statusCode, responseBytes = suite.createAList(userAUUID, inputAlist)
	suite.Equal(http.StatusCreated, statusCode)
	json.Unmarshal(responseBytes, &raw)
	alistUUID := raw["uuid"].(string)
	statusCode, responseBytes = suite.getList(userBUUID, alistUUID)
	suite.Equal(http.StatusForbidden, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.AclHttpAccessDeny, raw["message"].(string))
	raw = nil

	// Now we grant access for userBUUID.
	inputGrant := &HttpShareListWithUserInput{
		UserUUID:  userBUUID,
		AlistUUID: alistUUID,
		Action:    "grant",
	}
	a, _ := json.Marshal(inputGrant)
	statusCode, responseBytes = suite.shareAlist(userAUUID, string(a))
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(alistUUID, raw["alist_uuid"].(string))
	suite.Equal("grant", raw["action"].(string))
	suite.Equal(userBUUID, raw["user_uuid"].(string))
	raw = nil
	// Check to see if it worked by getting the list from above.
	statusCode, responseBytes = suite.getList(userBUUID, alistUUID)
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(alistUUID, raw["uuid"].(string))
	raw = nil

	inputRevoke := &HttpShareListWithUserInput{
		UserUUID:  userBUUID,
		AlistUUID: alistUUID,
		Action:    "revoke",
	}
	b, _ := json.Marshal(inputRevoke)
	statusCode, responseBytes = suite.shareAlist(userAUUID, string(b))
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(alistUUID, raw["alist_uuid"].(string))
	suite.Equal("revoke", raw["action"].(string))
	suite.Equal(userBUUID, raw["user_uuid"].(string))
	raw = nil
	// Check to see if it worked by getting the list from above.
	statusCode, responseBytes = suite.getList(userBUUID, alistUUID)
	suite.Equal(http.StatusForbidden, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.AclHttpAccessDeny, raw["message"].(string))
	raw = nil
}

func (suite *ApiSuite) shareAlist(userUUID string, input string) (statusCode int, responseBytes []byte) {
	user := &uuid.User{
		Uuid: userUUID,
	}

	req, rec := setupFakeEndpoint(http.MethodPost, "/v1/share/alist", input)
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)

	suite.NoError(env.V1ShareAlist(c))
	return rec.Code, rec.Body.Bytes()
}
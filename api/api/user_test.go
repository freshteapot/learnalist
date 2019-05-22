package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/labstack/echo/v4"
)

// @todo this needs faking of the actual database commands.

func (suite *ApiSuite) TestPostRegisterEmptyBody() {
	var raw map[string]interface{}
	statusCode, jsonBytes := suite.createNewUser("")
	suite.Equal(http.StatusBadRequest, statusCode)
	json.Unmarshal(jsonBytes, &raw)
	suite.Equal(i18n.ValidationUserRegister, raw["message"].(string))
}

func (suite *ApiSuite) TestPostRegisterNotValidJSON() {
	badInput := `{username:"chris", password:"test"}`
	var raw map[string]interface{}
	statusCode, jsonBytes := suite.createNewUser(badInput)
	suite.Equal(http.StatusBadRequest, statusCode)
	json.Unmarshal(jsonBytes, &raw)
	suite.Equal(i18n.ValidationUserRegister, raw["message"].(string))
}

func (suite *ApiSuite) TestPostRegisterNotValidPayload() {
	badInput := `{"username":"", "password":""}`
	var raw map[string]interface{}
	statusCode, jsonBytes := suite.createNewUser(badInput)
	suite.Equal(http.StatusBadRequest, statusCode)
	json.Unmarshal(jsonBytes, &raw)
	suite.Equal(i18n.ValidationUserRegister, raw["message"].(string))
}

func (suite *ApiSuite) TestPostRegisterValidPayload() {
	input := getValidUserRegisterInput("a")

	statusCode, jsonBytes := suite.createNewUser(input)
	suite.Equal(http.StatusCreated, statusCode)
	responseA := strings.TrimSpace(string(jsonBytes))
	// Check we get the same userid
	statusCode, jsonBytes = suite.createNewUser(input)
	suite.Equal(http.StatusOK, statusCode)
	responseB := strings.TrimSpace(string(jsonBytes))

	suite.Equal(responseA, responseB)
}

func (suite *ApiSuite) TestPostRegisterValidPayloadThenFake() {
	input := getValidUserRegisterInput("a")
	fake := `{"username":"iamusera", "password":"test123456"}`
	expectedFakeResponse := fmt.Sprintf(`{"message":"%s"}`, i18n.UserInsertUsernameExists)
	statusCode, _ := suite.createNewUser(input)
	suite.Equal(http.StatusCreated, statusCode)

	statusCode, jsonBytes := suite.createNewUser(fake)
	suite.Equal(http.StatusBadRequest, statusCode)
	response := strings.TrimSpace(string(jsonBytes))
	suite.Equal(response, expectedFakeResponse)
}

func (suite *ApiSuite) TestPostRegisterRepeat() {
	var statusCode int
	input := getValidUserRegisterInput("a")
	_, statusCode = suite.createNewUserWithSuccess(input)
	suite.Equal(http.StatusCreated, statusCode)

	_, statusCode = suite.createNewUserWithSuccess(input)
	suite.Equal(http.StatusOK, statusCode)
}

func (suite *ApiSuite) createNewUser(input string) (statusCode int, responseBytes []byte) {
	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/v1/register", input)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)
	suite.NoError(env.V1PostRegister(c))
	return rec.Code, rec.Body.Bytes()
}

func (suite *ApiSuite) createNewUserWithSuccess(input string) (uuid string, httpStatusCode int) {
	var raw map[string]interface{}
	statusCode, jsonBytes := suite.createNewUser(input)
	suite.Contains([]int{http.StatusOK, http.StatusCreated}, statusCode)
	json.Unmarshal(jsonBytes, &raw)
	user_uuid := raw["uuid"].(string)
	return user_uuid, statusCode
}

func getValidUserRegisterInput(which string) string {
	if which == "b" {
		return `{"username":"iamuserb", "password":"test123"}`
	}

	return `{"username":"iamusera", "password":"test123"}`
}

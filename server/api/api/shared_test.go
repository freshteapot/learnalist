package api

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (suite *ApiSuite) createNewUser(input string) (statusCode int, responseBytes []byte) {
	e := echo.New()
	req, rec := setupFakeEndpoint(http.MethodPost, "/api/v1/register", input)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)
	suite.NoError(m.V1PostRegister(c))
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

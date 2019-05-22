package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
)

func (suite *ApiSuite) TestPostLabel() {
	var raw map[string]interface{}
	var statusCode int
	var responseBytes []byte
	inputUserA := getValidUserRegisterInput("a")
	inputA := `{"label": "car"}`
	inputB := `{"label": "boat"}`
	inputC := `"bad data"`

	userUUID, _ := suite.createNewUserWithSuccess(inputUserA)

	statusCode, responseBytes = suite.postAlabel(userUUID, inputA)
	suite.Equal(http.StatusCreated, statusCode, "Label should have been created.")
	suite.Equal(`["car"]`, strings.TrimSpace(string(responseBytes)))

	statusCode, _ = suite.postAlabel(userUUID, inputA)
	suite.Equal(http.StatusOK, statusCode)

	statusCode, responseBytes = suite.postAlabel(userUUID, inputB)
	suite.Equal(http.StatusCreated, statusCode)
	suite.Equal(`["boat","car"]`, strings.TrimSpace(string(responseBytes)))

	statusCode, responseBytes = suite.postAlabel(userUUID, inputC)
	suite.Equal(http.StatusBadRequest, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal(i18n.PostUserLabelJSONFailure, raw["message"].(string))
}

func (suite *ApiSuite) TestGetUsersLabels() {
	var statusCode int
	var responseBytes []byte
	var response string

	inputUserA := getValidUserRegisterInput("a")
	userUUID, _ := suite.createNewUserWithSuccess(inputUserA)

	statusCode, responseBytes = suite.getLabels(userUUID)
	response = strings.TrimSpace(string(responseBytes))
	// Check it is an empty array
	suite.Equal("[]", response)

	// Post a label
	input := []string{
		`{"label": "car"}`,
		`{"label": "boat"}`,
		`{"label": "car"}`,
	}
	expectStatusCode := []int{201, 201, 200}
	for index, item := range input {
		statusCode, _ = suite.postAlabel(userUUID, item)
		suite.Equal(expectStatusCode[index], statusCode)
	}

	statusCode, responseBytes = suite.getLabels(userUUID)
	response = strings.TrimSpace(string(responseBytes))
	suite.Equal(`["boat","car"]`, response)
}

func (suite *ApiSuite) TestDeleteUsersLabels() {
	var req *http.Request
	var rec *httptest.ResponseRecorder
	var c echo.Context
	e := echo.New()
	user := uuid.NewUser()

	req, rec = setupFakeEndpoint(http.MethodDelete, "/v1/labels/car", "")
	c = e.NewContext(req, rec)
	c.Set("loggedInUser", user)
	suite.NoError(env.V1RemoveUserLabel(c))
	response := strings.TrimSpace(rec.Body.String())
	suite.Equal(`{"message":"Label car was removed."}`, response)
}

func (suite *ApiSuite) postAlabel(userUUID string, input string) (statusCode int, responseBytes []byte) {
	user := &uuid.User{
		Uuid: userUUID,
	}

	method := http.MethodPost
	uri := "/v1/labels"
	req, rec := setupFakeEndpoint(method, uri, input)
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	suite.NoError(env.V1PostUserLabel(c))
	return rec.Code, rec.Body.Bytes()
}

func (suite *ApiSuite) getLabels(userUUID string) (statusCode int, responseBytes []byte) {
	method := http.MethodGet
	uri := "/v1/labels/by/me"
	user := &uuid.User{
		Uuid: userUUID,
	}
	req, rec := setupFakeEndpoint(method, uri, "")
	e := echo.New()
	c := e.NewContext(req, rec)
	c.Set("loggedInUser", *user)
	suite.NoError(env.V1GetUserLabels(c))
	return rec.Code, rec.Body.Bytes()
}

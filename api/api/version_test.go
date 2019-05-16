package api

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (suite *ApiSuite) TestGetVersion() {
	var raw map[string]interface{}
	statusCode, responseBytes := suite.getVersion()
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal("n/a", raw["gitHash"].(string))
	suite.Equal("n/a", raw["gitDate"].(string))
	suite.Equal("n/a", raw["version"].(string))
}

func (suite *ApiSuite) getVersion() (statusCode int, responseBytes []byte) {
	method := http.MethodGet
	uri := "/version"

	req, rec := setupFakeEndpoint(method, uri, "")
	e := echo.New()
	c := e.NewContext(req, rec)
	suite.NoError(env.GetVersion(c))
	return rec.Code, rec.Body.Bytes()
}

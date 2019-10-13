package api_test

/*
func (suite *ApiSuite) TestGetVersion() {
	var raw map[string]interface{}
	statusCode, responseBytes := suite.getVersion()
	suite.Equal(http.StatusOK, statusCode)
	json.Unmarshal(responseBytes, &raw)
	suite.Equal("n/a", raw["gitHash"].(string))
	suite.Equal("n/a", raw["gitDate"].(string))
	suite.Equal("n/a", raw["version"].(string))
	suite.Equal("https://github.com/freshteapot/learnalist-api/commit/n_a", raw["url"].(string))
}

func (suite *ApiSuite) getVersion() (statusCode int, responseBytes []byte) {
	method := http.MethodGet
	uri := "/api/v1/version"

	req, rec := setupFakeEndpoint(method, uri, "")
	e := echo.New()
	c := e.NewContext(req, rec)
	suite.NoError(m.V1GetVersion(c))
	return rec.Code, rec.Body.Bytes()
}
*/

//2, as I am being really lazy :(, once all moved over to ginkgo remove.
import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Version endpoint", func() {
	It("Simple response", func() {
		input := ""
		method := http.MethodGet
		uri := "/api/v1/version"
		req, rec := setupFakeEndpoint(method, uri, input)

		e := echo.New()
		c := e.NewContext(req, rec)

		m.V1GetVersion(c)
		Expect(rec.Code).To(Equal(http.StatusOK))

		var raw map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &raw)
		Expect(raw["gitHash"].(string)).To(Equal("n/a"))
		Expect(raw["gitDate"].(string)).To(Equal("n/a"))
		Expect(raw["version"].(string)).To(Equal("n/a"))
		Expect(raw["url"].(string)).To(Equal("https://github.com/freshteapot/learnalist-api/commit/n_a"))
	})
})

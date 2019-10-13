package api_test

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

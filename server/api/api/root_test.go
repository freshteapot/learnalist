package api_test

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Root endpoint", func() {
	It("Simple response", func() {
		input := ""
		method := http.MethodGet
		uri := "/api/v1/"
		req, rec := setupFakeEndpoint(method, uri, input)

		e := echo.New()
		c := e.NewContext(req, rec)

		m.V1GetRoot(c)
		Expect(rec.Code).To(Equal(http.StatusOK))
		testutils.CheckMessageResponseFromResponseRecorder(rec, "1, 2, 3. Lets go!")
	})
})

package authenticate_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	authenticate "github.com/freshteapot/learnalist-api/server/api/authenticate"
)

var _ = Describe("Auth", func() {
	var (
		e   *echo.Echo
		req *http.Request
		rec *httptest.ResponseRecorder
		c   echo.Context
	)

	BeforeEach(func() {
		e = echo.New()
		req = httptest.NewRequest(http.MethodPost, "/", nil)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
	})

	It("return false by default", func() {
		skip := authenticate.Skip(c)
		Expect(skip).To(Equal(false))
	})

	It("reject based on method not supported", func() {
		req.Method = "fake"
		skip := authenticate.Skip(c)
		Expect(skip).To(Equal(false))
	})

	It("skip for /", func() {
		req.Method = http.MethodGet
		req.URL, _ = url.ParseRequestURI("/api/v1/")
		skip := authenticate.Skip(c)
		Expect(skip).To(Equal(true))
	})

	It("skip for /version", func() {
		req.Method = http.MethodGet
		req.URL, _ = url.ParseRequestURI("/api/v1/version")
		skip := authenticate.Skip(c)
		Expect(skip).To(Equal(true))
	})

	It("skip for callback requests with prefix /oauth/", func() {
		req.Method = http.MethodGet
		req.URL, _ = url.ParseRequestURI("/api/v1/oauth/callback-from-somehwere")
		skip := authenticate.Skip(c)
		Expect(skip).To(Equal(true))
	})

	It("skip for insecure user registration", func() {
		req.Method = http.MethodPost
		req.URL, _ = url.ParseRequestURI("/api/v1/user/register")
		skip := authenticate.Skip(c)
		Expect(skip).To(Equal(true))
	})
})

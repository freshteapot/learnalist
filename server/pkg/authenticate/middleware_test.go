package authenticate_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	authenticate "github.com/freshteapot/learnalist-api/server/pkg/authenticate"
)

var _ = Describe("Middleware", func() {
	var (
		e    *echo.Echo
		req  *http.Request
		rec  *httptest.ResponseRecorder
		c    echo.Context
		next func(c echo.Context) (err error)
	)

	BeforeEach(func() {
		e = echo.New()
		req = httptest.NewRequest(http.MethodPost, "/", nil)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		next = func(c echo.Context) (err error) {
			return c.NoContent(http.StatusOK)
		}
	})

	It("Skip", func() {
		config := authenticate.Config{
			Skip: func(c echo.Context) bool {
				return true
			},
			LookupBasic: func(username string, hash string) (string, error) {
				return "", nil
			},
			LookupBearer: func(token string) (string, error) {
				return "", nil
			},
		}
		middleware := authenticate.Auth(config)
		middleware(next)(c)
		Expect(rec.Code).To(Equal(http.StatusOK))
	})

	It("No authorization header", func() {
		config := authenticate.Config{
			Skip: func(c echo.Context) bool {
				return false
			},
			LookupBasic: func(username string, hash string) (string, error) {
				return "", nil
			},
			LookupBearer: func(token string) (string, error) {
				return "", nil
			},
		}
		middleware := authenticate.Auth(config)
		result := middleware(next)(c).(*echo.HTTPError)
		Expect(result.Code).To(Equal(http.StatusForbidden))
	})

	When("Authorization header is present", func() {
		It("Empty", func() {
			req.Header.Set("Authorization", "")
			config := authenticate.Config{
				Skip: func(c echo.Context) bool {
					return false
				},
				LookupBasic: func(username string, hash string) (string, error) {
					return "", nil
				},
				LookupBearer: func(token string) (string, error) {
					return "", nil
				},
			}
			middleware := authenticate.Auth(config)
			result := middleware(next)(c).(*echo.HTTPError)
			Expect(result.Code).To(Equal(http.StatusForbidden))
		})

		It("not supported type", func() {
			req.Header.Set("Authorization", "a b")
			config := authenticate.Config{
				Skip: func(c echo.Context) bool {
					return false
				},
				LookupBasic: func(username string, hash string) (string, error) {
					return "", nil
				},
				LookupBearer: func(token string) (string, error) {
					return "", nil
				},
			}
			middleware := authenticate.Auth(config)
			result := middleware(next)(c).(*echo.HTTPError)
			Expect(result.Code).To(Equal(http.StatusForbidden))
		})

		When("bearer", func() {
			It("error on looking up", func() {
				req.Header.Set("Authorization", "Bearer b")
				config := authenticate.Config{
					Skip: func(c echo.Context) bool {
						return false
					},
					LookupBasic: func(username string, hash string) (string, error) {
						return "", nil
					},
					LookupBearer: func(token string) (string, error) {
						return "", errors.New("yo")
					},
				}
				middleware := authenticate.Auth(config)
				result := middleware(next)(c).(*echo.HTTPError)
				Expect(result.Code).To(Equal(http.StatusUnauthorized))
				Expect(strings.HasPrefix(rec.Header().Get("WWW-Authenticate"), "bearer")).To(BeTrue())
			})

			It("success, user found and added to the context", func() {
				req.Header.Set("Authorization", "Bearer b")
				config := authenticate.Config{
					Skip: func(c echo.Context) bool {
						return false
					},
					LookupBasic: func(username string, hash string) (string, error) {
						return "", nil
					},
					LookupBearer: func(token string) (string, error) {
						return "test", nil
					},
				}
				middleware := authenticate.Auth(config)
				middleware(next)(c)
				Expect(c.Get("loggedInUser").(uuid.User).Uuid).To(Equal("test"))
			})
		})

		When("basic", func() {
			It("error, not valid basic auth format", func() {
				req.Header.Set("Authorization", "Basic fake")
				config := authenticate.Config{
					Skip: func(c echo.Context) bool {
						return false
					},
					LookupBasic: func(username string, hash string) (string, error) {
						return "", nil
					},
					LookupBearer: func(token string) (string, error) {
						return "", nil
					},
				}
				middleware := authenticate.Auth(config)
				result := middleware(next)(c).(*echo.HTTPError)
				Expect(result.Code).To(Equal(http.StatusUnauthorized))
				Expect(strings.HasPrefix(rec.Header().Get("WWW-Authenticate"), "basic")).To(BeTrue())
			})

			It("error, failed to find the user", func() {
				req.Header.Set("Authorization", "Basic YTpi")
				config := authenticate.Config{
					Skip: func(c echo.Context) bool {
						return false
					},
					LookupBasic: func(username string, hash string) (string, error) {
						return "", errors.New("you must fail")
					},
					LookupBearer: func(token string) (string, error) {
						return "", nil
					},
				}
				middleware := authenticate.Auth(config)
				result := middleware(next)(c).(*echo.HTTPError)
				Expect(result.Code).To(Equal(http.StatusUnauthorized))
				Expect(strings.HasPrefix(rec.Header().Get("WWW-Authenticate"), "basic")).To(BeTrue())
			})

			It("success, user found and added to the context", func() {
				req.Header.Set("Authorization", "Basic YTpi")
				config := authenticate.Config{
					Skip: func(c echo.Context) bool {
						return false
					},
					LookupBasic: func(username string, hash string) (string, error) {
						return "test", nil
					},
					LookupBearer: func(token string) (string, error) {
						return "", nil
					},
				}
				middleware := authenticate.Auth(config)
				middleware(next)(c)
				Expect(c.Get("loggedInUser").(uuid.User).Uuid).To(Equal("test"))
			})
		})
	})
})

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func init() {
	resetDatabase()
}

func TestGetRoot(t *testing.T) {
	resetDatabase()
	expected := `{"message":"1, 2, 3. Lets go!"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, env.GetRoot(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		assert.Equal(t, expected, response)
	}

}

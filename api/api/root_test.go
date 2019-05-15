package api

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/labstack/echo/v4"
)

func (suite *ApiSuite) TestGetRoot() {
	expected := `{"message":"1, 2, 3. Lets go!"}`
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if suite.NoError(env.GetRoot(c)) {
		suite.Equal(http.StatusOK, rec.Code)
		response := strings.TrimSpace(rec.Body.String())
		suite.Equal(expected, response)
	}

}

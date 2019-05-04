package api

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/freshteapot/learnalist-api/api/api/models"
	"github.com/labstack/echo"
)

var env = Env{
	Port:         9090,
	DatabaseName: "./test.db",
}

func resetDatabase() {
	db, err := models.NewTestDB()
	if err != nil {
		log.Panic(err)
	}

	env.Datastore = &models.DAL{
		Db: db,
	}
}

func setupFakeEndpoint(method string, uri string, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	return req, rec
}

// @TODO
/*
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/freshteapot/learnalist-api/api/api/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/engine/standard"
	"github.com/stretchr/testify/assert"
)

type TestDAL struct {
	Db *sql.DB
}

var (
	env = &Env{
		Datastore: &TestDAL{},
	}
)

func (dal *TestDAL) GetListsBy(uuid string) ([]*models.Alist, error) {
	var items = []*models.Alist{}
	data := `[{"data":["a","b"],"info":{"title":"I am a list","type":"v1"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"},{"data":{"car":"bil","water":"vann"},"info":{"title":"I am a list with items","type":"v2"},"uuid":"efeb4a6e-9a03-5aff-b46d-7f2ba1d7e7f9"}]`
	err := json.Unmarshal([]byte(data), &items)
	if err != nil {
		fmt.Println(err)
	}
	return items, nil
}
func (dal *TestDAL) GetAlist(uuid string) (*models.Alist, error) {
	var item *models.Alist
	return item, nil
}
func (dal *TestDAL) PostAlist(interface{}) (*models.Alist, error) {
	var item *models.Alist
	return item, nil
}
func (dal *TestDAL) UpdateAlist(interface{}) (*models.Alist, error) {
	var item *models.Alist
	return item, nil
}
func (dal *TestDAL) CreateDBStructure() {

}

var _ models.Datastore = (*TestDAL)(nil)

func TestRoot(t *testing.T) {
	expectedResponse := `{"message":"1, 2, 3. Lets go!"}`
	e := echo.New()
	req := new(http.Request)
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))
	c.SetPath("/")

	if assert.NoError(t, env.GetRoot(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expectedResponse, rec.Body.String())
	}
}

func TestGetListBy(t *testing.T) {
	expectedResponse := `[{"data":["a","b"],"info":{"title":"I am a list","type":"v1"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"},{"data":{"car":"bil","water":"vann"},"info":{"title":"I am a list with items","type":"v2"},"uuid":"efeb4a6e-9a03-5aff-b46d-7f2ba1d7e7f9"}]`

	e := echo.New()
	req := new(http.Request)
	rec := httptest.NewRecorder()
	c := e.NewContext(standard.NewRequest(req, e.Logger()), standard.NewResponse(rec, e.Logger()))
	c.SetPath("/alist/by/:uuid")
	c.SetParamNames("uuid")
	c.SetParamValues("me")

	if assert.NoError(t, env.GetListsBy(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		fmt.Println(rec.Body.String())
		assert.Equal(t, expectedResponse, rec.Body.String())
	}
}
*/

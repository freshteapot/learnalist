package testutils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/gomega"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func CleanEchoResponseFromResponseRecorder(rec *httptest.ResponseRecorder) string {
	return CleanEchoResponseFromByte(rec.Body.Bytes())
}

func CleanEchoResponseFromByte(data []byte) string {
	return strings.TrimSuffix(string(data), "\n")
}

func CheckMessageResponse(response api.HTTPResponse, expect string) {
	var obj api.HTTPResponseMessage
	json.Unmarshal(response.Body, &obj)
	Expect(obj.Message).To(Equal(expect))
}

func CheckMessageResponseFromReader(body io.Reader, expect string) {
	data, err := ioutil.ReadAll(body)
	Expect(err).To(BeNil())

	var response api.HTTPResponseMessage
	json.Unmarshal(data, &response)
	Expect(response.Message).To(Equal(expect))
}

func CheckMessageResponseFromResponseRecorder(rec *httptest.ResponseRecorder, expect string) {
	reader := bytes.NewReader(rec.Body.Bytes())
	CheckMessageResponseFromReader(reader, expect)
}

func ToHttpResponse(response *http.Response, err error) (api.HTTPResponse, error) {
	// TODO what todo with err
	var obj api.HTTPResponse
	defer response.Body.Close()
	obj.StatusCode = response.StatusCode
	buf, _ := ioutil.ReadAll(response.Body)
	obj.Body = buf
	return obj, err
}

func SetupJSONEndpoint(method string, uri string, body string) (*http.Request, *httptest.ResponseRecorder) {
	var input io.Reader
	if body != "" {
		input = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, uri, input)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	return req, rec
}

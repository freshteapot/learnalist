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
	. "github.com/onsi/gomega"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func CleanEchoResponseFromResponseRecorder(rec *httptest.ResponseRecorder) string {
	return strings.TrimSuffix(string(rec.Body.Bytes()), "\n")
}

func CleanEchoResponseFromByte(data []byte) string {
	return strings.TrimSuffix(string(data), "\n")
}

func CheckMessageResponse(response api.HttpResponse, expect string) {
	var obj api.HttpResponseMessage
	json.Unmarshal(response.Body, &obj)
	Expect(obj.Message).To(Equal(expect))
}

func CheckMessageResponseFromReader(body io.Reader, expect string) {
	data, err := ioutil.ReadAll(body)
	Expect(err).To(BeNil())

	var response api.HttpResponseMessage
	json.Unmarshal(data, &response)
	Expect(response.Message).To(Equal(expect))
}

func CheckMessageResponseFromResponseRecorder(rec *httptest.ResponseRecorder, expect string) {
	reader := bytes.NewReader(rec.Body.Bytes())
	CheckMessageResponseFromReader(reader, expect)
}

func ToHttpResponse(response *http.Response, err error) (api.HttpResponse, error) {
	// TODO what todo with err
	var obj api.HttpResponse
	defer response.Body.Close()
	obj.StatusCode = response.StatusCode
	buf, _ := ioutil.ReadAll(response.Body)
	obj.Body = buf
	return obj, err
}

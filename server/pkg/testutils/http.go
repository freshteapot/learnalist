package testutils

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

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

func CheckMessageResponse(response api.HttpResponse, expect string) {
	var obj api.HttpResponseMessage
	json.Unmarshal(response.Body, &obj)
	Expect(obj.Message).To(Equal(expect))
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

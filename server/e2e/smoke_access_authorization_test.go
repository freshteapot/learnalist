package e2e_test

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/e2e"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("access authorization issues73", func() {

		var (
			req      *http.Request
			response *http.Response
			err      error
		)

		assert := assert.New(GinkgoT())
		learnalistClient := e2e.NewClient(server)
		url := fmt.Sprintf("%s/api/v1/alist/by/me", server)

		fmt.Println("> Request with wrong authorization header")
		req, err = http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "Fake a")
		req.Header.Set("Content-Type", "application/json")
		response, err = learnalistClient.RawRequest(req)
		assert.NoError(err)
		assert.Equal(response.StatusCode, http.StatusForbidden)

		fmt.Println("> Request with empty authorization header")
		req, err = http.NewRequest(http.MethodGet, url, nil)
		req.Header.Set("Authorization", "")
		req.Header.Set("Content-Type", "application/json")
		response, err = learnalistClient.RawRequest(req)
		assert.NoError(err)
		assert.Equal(response.StatusCode, http.StatusForbidden)

		fmt.Println("> Request without setting the authorization header, doesnt crash the app")
		req, err = http.NewRequest(http.MethodGet, url, nil)
		assert.NoError(err)
		response, err = learnalistClient.RawRequest(req)
		assert.NoError(err)
		assert.Equal(response.StatusCode, http.StatusForbidden)
	})
})

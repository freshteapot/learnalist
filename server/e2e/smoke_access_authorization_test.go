package e2e_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/freshteapot/learnalist-api/server/e2e"
	"github.com/stretchr/testify/assert"
)

// TestAccessAuthorizationIssues73 make sure the issue doesnt return
// https://github.com/freshteapot/learnalist-api/issues/73
func TestAccessAuthorizationIssues73(t *testing.T) {
	var (
		req      *http.Request
		response *http.Response
		err      error
	)

	assert := assert.New(t)
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
}

package e2e_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/e2e"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/stretchr/testify/assert"
)

func TestListAccess(t *testing.T) {
	var httpResponse e2e.HttpResponse
	var messageResponse e2e.MessageResponse
	var err error
	var aList alist.Alist

	assert := assert.New(t)
	learnalistClient := e2e.NewClient(server)
	fmt.Println("> Register users owner and reader")
	userInfoOwner := learnalistClient.Register(usernameOwner, password)
	userInfoReader := learnalistClient.Register(usernameReader, password)
	fmt.Println("> Create a list")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
	fmt.Println("> Read user is unable to view it")
	httpResponse = learnalistClient.GetListByUUIDV1(userInfoReader, aList.Uuid)
	assert.Equal(httpResponse.StatusCode, http.StatusForbidden)
	fmt.Println("> List is now public")
	messageResponse = learnalistClient.SetListShareV1(userInfoOwner, aList.Uuid, aclKeys.SharedWithPublic)
	assert.Equal(messageResponse.Message, "List is now public")
	httpResponse = learnalistClient.GetListByUUIDV1(userInfoReader, aList.Uuid)
	assert.Equal(httpResponse.StatusCode, http.StatusOK)
	fmt.Println("> Check access to the html pages as well")
	fmt.Println(">> Too fast for the rebuild")
	httpResponse, err = learnalistClient.GetAlistHtml(userInfoReader, aList.Uuid)
	assert.NoError(err)
	assert.Equal(httpResponse.StatusCode, http.StatusOK)
	assert.True(strings.Contains(string(httpResponse.Body), "Please refresh"))
	fmt.Println(">> A human wait, should have contents")
	time.Sleep(100 * time.Millisecond)

	httpResponse, err = learnalistClient.GetAlistHtml(userInfoReader, aList.Uuid)
	assert.NoError(err)
	assert.Equal(httpResponse.StatusCode, http.StatusOK)
	assert.NoError(err)
	assert.Equal(httpResponse.StatusCode, http.StatusOK)
	assert.True(strings.Contains(string(httpResponse.Body), "<title>Days of the Week</title>"))

	fmt.Println("> Share the list with friends only")
	messageResponse = learnalistClient.SetListShareV1(userInfoOwner, aList.Uuid, aclKeys.SharedWithFriends)
	assert.Equal(messageResponse.Message, "List is now private to the owner and those granted access")

	fmt.Println("> Confirm the other user cant access the list via html or the api")
	httpResponse, _ = learnalistClient.GetAlistHtml(userInfoReader, aList.Uuid)
	assert.NoError(err)
	assert.Equal(httpResponse.StatusCode, http.StatusForbidden)
	assert.True(strings.Contains(string(httpResponse.Body), "<title>A list: access denied for this list</title>"))
	httpResponse = learnalistClient.GetListByUUIDV1(userInfoReader, aList.Uuid)
	assert.Equal(httpResponse.StatusCode, http.StatusForbidden)
	assert.Equal(cleanEchoJSONResponse(httpResponse.Body), `{"message":"Access Denied"}`)

	fmt.Println("> Set the other user to be able to read the list")
	httpResponse, err = learnalistClient.ShareReadAcessV1(userInfoOwner, aList.Uuid, userInfoReader.Uuid, aclKeys.ActionGrant)
	assert.NoError(err)
	assert.Equal(httpResponse.StatusCode, http.StatusOK)

	fmt.Println("> Confirm the other user can access the list via html or the api")
	httpResponse = learnalistClient.GetListByUUIDV1(userInfoReader, aList.Uuid)
	assert.Equal(httpResponse.StatusCode, http.StatusOK)
	httpResponse, err = learnalistClient.GetAlistHtml(userInfoReader, aList.Uuid)
	assert.NoError(err)
	assert.Equal(httpResponse.StatusCode, http.StatusOK)
	assert.True(strings.Contains(string(httpResponse.Body), "<title>Days of the Week</title>"))
}
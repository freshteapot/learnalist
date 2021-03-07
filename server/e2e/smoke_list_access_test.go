package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Smoke list access", func() {
	var (
		userInfoOwner, userInfoReader openapi.HttpUserLoginResponse
		authOwner, authReader         context.Context
	)

	AfterEach(func() {
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
		openapiClient.DeleteUser(authReader, userInfoReader.UserUuid)
	})

	It("list access", func() {

		var httpResponse api.HTTPResponse
		var messageResponse api.HTTPResponseMessage
		var err error
		var aList alist.Alist

		assert := assert.New(GinkgoT())
		fmt.Println("> Register users owner and reader")
		authOwner, userInfoOwner = RegisterAndLogin(openapiClient.API)
		authReader, userInfoReader = RegisterAndLogin(openapiClient.API)

		fmt.Println("> Create a list")
		aList, _ = e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))

		fmt.Println("> Read user is unable to view it")
		httpResponse = e2eClient.GetListByUUIDV1(userInfoReader, aList.Uuid)
		assert.Equal(httpResponse.StatusCode, http.StatusForbidden)
		fmt.Println("> List is now public")
		messageResponse = e2eClient.SetListShareV1(userInfoOwner, aList.Uuid, aclKeys.SharedWithPublic)

		assert.Equal(messageResponse.Message, "List is now public")

		httpResponse = e2eClient.GetListByUUIDV1(userInfoReader, aList.Uuid)
		assert.Equal(httpResponse.StatusCode, http.StatusOK)

		fmt.Println("> Check access to the html pages as well")
		fmt.Println(">> Too fast for the rebuild")
		httpResponse, err = e2eClient.GetAlistHtml(userInfoReader, aList.Uuid)

		assert.NoError(err)
		assert.Equal(httpResponse.StatusCode, http.StatusNotFound)
		fmt.Println(">> A human wait, should have contents")
		// TODO I wonder how to make this more reliable
		time.Sleep(1000 * time.Millisecond)

		httpResponse, err = e2eClient.GetAlistHtml(userInfoReader, aList.Uuid)
		assert.NoError(err)
		assert.Equal(httpResponse.StatusCode, http.StatusOK)
		//fmt.Println(aList.Uuid)
		//fmt.Println(string(httpResponse.Body))
		assert.True(strings.Contains(string(httpResponse.Body), "<title>Days of the Week</title>"))

		fmt.Println("> Share the list with friends only")
		messageResponse = e2eClient.SetListShareV1(userInfoOwner, aList.Uuid, aclKeys.SharedWithFriends)
		assert.Equal(messageResponse.Message, "List is now private to the owner and those granted access")

		fmt.Println("> Confirm the other user cant access the list via html or the api")
		httpResponse, _ = e2eClient.GetAlistHtml(userInfoReader, aList.Uuid)
		assert.NoError(err)
		assert.Equal(httpResponse.StatusCode, http.StatusForbidden)
		assert.True(strings.Contains(string(httpResponse.Body), "<title>A list: access denied for this list</title>"))
		httpResponse = e2eClient.GetListByUUIDV1(userInfoReader, aList.Uuid)
		assert.Equal(httpResponse.StatusCode, http.StatusForbidden)
		assert.Equal(testutils.CleanEchoResponseFromByte(httpResponse.Body), `{"message":"Access Denied"}`)

		fmt.Println("> Set the other user to be able to read the list")
		httpResponse, err = e2eClient.ShareReadAcessV1(userInfoOwner, aList.Uuid, userInfoReader.UserUuid, aclKeys.ActionGrant)
		assert.NoError(err)
		assert.Equal(httpResponse.StatusCode, http.StatusOK)

		fmt.Println("> Confirm the other user can access the list via html or the api")
		httpResponse = e2eClient.GetListByUUIDV1(userInfoReader, aList.Uuid)
		assert.Equal(httpResponse.StatusCode, http.StatusOK)
		httpResponse, err = e2eClient.GetAlistHtml(userInfoReader, aList.Uuid)
		assert.NoError(err)
		assert.Equal(httpResponse.StatusCode, http.StatusOK)
		assert.True(strings.Contains(string(httpResponse.Body), "<title>Days of the Week</title>"))
	})
})

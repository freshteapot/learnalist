package e2e_test

import (
	"context"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Testing Smoke test 1", func() {
	var (
		userInfoOwner, userInfoReader openapi.HttpUserLoginResponse
		authOwner, authReader         context.Context
	)

	AfterEach(func() {
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
		openapiClient.DeleteUser(authReader, userInfoReader.UserUuid)
	})

	It("share public", func() {
		var httpResponse api.HTTPResponse
		var messageResponse api.HTTPResponseMessage

		assert := assert.New(GinkgoT())

		authOwner, userInfoOwner = RegisterAndLogin(openapiClient.API)
		authReader, userInfoReader = RegisterAndLogin(openapiClient.API)

		listInfo, _ := e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))

		httpResponse = e2eClient.GetListByUUIDV1(userInfoReader, listInfo.Uuid)
		assert.Equal(httpResponse.StatusCode, 403)
		messageResponse = e2eClient.SetListShareV1(userInfoOwner, listInfo.Uuid, "public")
		assert.Equal(messageResponse.Message, "List is now public")
		httpResponse = e2eClient.GetListByUUIDV1(userInfoReader, listInfo.Uuid)
		assert.Equal(httpResponse.StatusCode, 200)
		messageResponse = e2eClient.SetListShareV1(userInfoOwner, listInfo.Uuid, "friends")
		assert.Equal(messageResponse.Message, "List is now private to the owner and those granted access")
		httpResponse = e2eClient.GetListByUUIDV1(userInfoReader, listInfo.Uuid)
		assert.Equal(httpResponse.StatusCode, 403)

		for j := 0; j <= 10; j++ {
			e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		}

		//for j := 0; j <= 20; j++ {
		//	go func() {
		//		e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		//	}()
		//}

	})

	It("share private", func() {

		var httpResponse api.HTTPResponse
		assert := assert.New(GinkgoT())

		authOwner, userInfoOwner = RegisterAndLogin(openapiClient.API)
		authReader, userInfoReader = RegisterAndLogin(openapiClient.API)

		listInfo, _ := e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		httpResponse = e2eClient.GetListByUUIDV1(userInfoReader, listInfo.Uuid)
		assert.Equal(httpResponse.StatusCode, http.StatusForbidden)
	})
})

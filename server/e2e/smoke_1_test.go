package e2e_test

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/api"
	"github.com/freshteapot/learnalist-api/server/e2e"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Testing Smoke test 1", func() {
	It("share public", func() {

		var httpResponse e2e.HttpResponse
		var messageResponse api.HttpResponseMessage

		assert := assert.New(GinkgoT())
		learnalistClient := e2e.NewClient(server)
		userInfoOwner := learnalistClient.Register(usernameOwner, password)
		fmt.Println(userInfoOwner.Uuid)
		userInfoReader := learnalistClient.Register(usernameReader, password)
		fmt.Println(userInfoReader.Uuid)
		listInfo, _ := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		fmt.Println(listInfo.Uuid)

		httpResponse = learnalistClient.GetListByUUIDV1(userInfoReader, listInfo.Uuid)
		assert.Equal(httpResponse.StatusCode, 403)
		messageResponse = learnalistClient.SetListShareV1(userInfoOwner, listInfo.Uuid, "public")
		assert.Equal(messageResponse.Message, "List is now public")
		httpResponse = learnalistClient.GetListByUUIDV1(userInfoReader, listInfo.Uuid)
		assert.Equal(httpResponse.StatusCode, 200)
		messageResponse = learnalistClient.SetListShareV1(userInfoOwner, listInfo.Uuid, "friends")
		assert.Equal(messageResponse.Message, "List is now private to the owner and those granted access")
		httpResponse = learnalistClient.GetListByUUIDV1(userInfoReader, listInfo.Uuid)
		assert.Equal(httpResponse.StatusCode, 403)

		for j := 0; j <= 10; j++ {
			learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		}

		for j := 0; j <= 100; j++ {
			go func() {
				learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
			}()
		}
	})
	It("share private", func() {

		var httpResponse e2e.HttpResponse
		assert := assert.New(GinkgoT())
		learnalistClient := e2e.NewClient(server)

		userInfoOwner := learnalistClient.Register(usernameOwner, password)
		userInfoReader := learnalistClient.Register(usernameReader, password)
		listInfo, _ := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		httpResponse = learnalistClient.GetListByUUIDV1(userInfoReader, listInfo.Uuid)
		assert.Equal(httpResponse.StatusCode, http.StatusForbidden)
	})
})

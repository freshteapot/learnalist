package e2e_test

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Testing Smoke test 2", func() {
	It("share public2", func() {

		var messageResponse api.HTTPResponseMessage

		assert := assert.New(GinkgoT())

		_, userInfoOwner := RegisterAndLogin(openapiClient.API)

		listInfo, _ := e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		messageResponse = e2eClient.SetListShareV1(userInfoOwner, listInfo.Uuid, "public")
		assert.Equal(messageResponse.Message, "List is now public")
		fmt.Println(fmt.Sprintf("http://localhost:1234/alist/%s.html", listInfo.Uuid))
	})
})

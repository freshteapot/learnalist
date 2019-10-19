package e2e_test

import (
	"fmt"
	"testing"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/e2e"
	"github.com/stretchr/testify/assert"
)

func TestSharePublic(t *testing.T) {
	var httpResponse e2e.HttpResponse
	var messageResponse e2e.MessageResponse

	assert := assert.New(t)
	learnalistClient := e2e.NewClient(server)

	userInfoOwner := learnalistClient.Register(usernameOwner, password)
	fmt.Println(userInfoOwner.Uuid)
	userInfoReader := learnalistClient.Register(usernameReader, password)
	fmt.Println(userInfoReader.Uuid)
	listInfo, _ := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
	fmt.Println(listInfo.Uuid)

	httpResponse = learnalistClient.GetListByUUID(userInfoReader, listInfo.Uuid)
	assert.Equal(httpResponse.StatusCode, 403)
	messageResponse = learnalistClient.SetListShare(userInfoOwner, listInfo.Uuid, "public")
	assert.Equal(messageResponse.Message, "List is now public")
	httpResponse = learnalistClient.GetListByUUID(userInfoReader, listInfo.Uuid)
	assert.Equal(httpResponse.StatusCode, 200)
	messageResponse = learnalistClient.SetListShare(userInfoOwner, listInfo.Uuid, "friends")
	assert.Equal(messageResponse.Message, "List is now private to the owner and those granted access")
	httpResponse = learnalistClient.GetListByUUID(userInfoReader, listInfo.Uuid)
	assert.Equal(httpResponse.StatusCode, 403)
	// Currently it doesnt handle too many requests
	for j := 0; j <= 10; j++ {
		learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
	}
	/*
		for j := 0; j <= 100; j++ {
			go func() {
				learnalistClient.PostListV1(userInfoOwner, inputAlistV1)
			}()
		}
	*/
}

func TestSharePrivate(t *testing.T) {
	var httpResponse e2e.HttpResponse
	assert := assert.New(t)
	learnalistClient := e2e.NewClient(server)

	userInfoOwner := learnalistClient.Register(usernameOwner, password)
	userInfoReader := learnalistClient.Register(usernameReader, password)
	listInfo, _ := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
	httpResponse = learnalistClient.GetListByUUID(userInfoReader, listInfo.Uuid)
	assert.Equal(httpResponse.StatusCode, 403)
}

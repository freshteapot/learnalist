package e2e_test

import (
	"fmt"
	"testing"

	"github.com/freshteapot/learnalist-api/server/e2e"
	"github.com/stretchr/testify/assert"
)

var usernameOwner = "iamchris"
var password = "test123"
var usernameReader = "iamusera"

var server = "http://127.0.0.1:1234"
var inputAlistV1 = `
{
  "data": [
      "monday",
      "tuesday",
      "wednesday",
      "thursday",
      "friday",
      "saturday",
      "sunday"
  ],
  "info": {
      "title": "Days of the Week",
      "type": "v1"
  }
}
`

func TestSharePublic(t *testing.T) {
	var httpResponse e2e.HttpResponse
	var messageResponse e2e.MessageResponse

	assert := assert.New(t)
	learnalistClient := e2e.NewClient(server)

	userInfoOwner := learnalistClient.Register(usernameOwner, password)
	fmt.Println(userInfoOwner.Uuid)
	userInfoReader := learnalistClient.Register(usernameReader, password)
	fmt.Println(userInfoReader.Uuid)
	listInfo := learnalistClient.PostListV1(userInfoOwner, inputAlistV1)
	fmt.Println(listInfo.Uuid)

	httpResponse = learnalistClient.GetListByUuID(userInfoReader, listInfo.Uuid)
	assert.Equal(httpResponse.StatusCode, 403)
	messageResponse = learnalistClient.SetListShare(userInfoOwner, listInfo.Uuid, "public")
	assert.Equal(messageResponse.Message, "List is now public")
	httpResponse = learnalistClient.GetListByUuID(userInfoReader, listInfo.Uuid)
	assert.Equal(httpResponse.StatusCode, 200)
	messageResponse = learnalistClient.SetListShare(userInfoOwner, listInfo.Uuid, "friends")
	assert.Equal(messageResponse.Message, "List is now private to the owner and those granted access")
	httpResponse = learnalistClient.GetListByUuID(userInfoReader, listInfo.Uuid)
	assert.Equal(httpResponse.StatusCode, 403)
	// Currently it doesnt handle too many requests
	for j := 0; j <= 10; j++ {
		learnalistClient.PostListV1(userInfoOwner, inputAlistV1)
	}

	for j := 0; j <= 200; j++ {
		go func() {
			learnalistClient.PostListV1(userInfoOwner, inputAlistV1)
		}()
	}

}

func TestSharePrivate(t *testing.T) {
	var httpResponse e2e.HttpResponse
	assert := assert.New(t)
	learnalistClient := e2e.NewClient(server)

	userInfoOwner := learnalistClient.Register(usernameOwner, password)
	userInfoReader := learnalistClient.Register(usernameReader, password)
	listInfo := learnalistClient.PostListV1(userInfoOwner, inputAlistV1)
	fmt.Println(listInfo)
	httpResponse = learnalistClient.GetListByUuID(userInfoReader, listInfo.Uuid)
	assert.Equal(httpResponse.StatusCode, 403)
	//httpResponse = learnalistClient.GetListByUuID(userInfoOwner, listInfo.Uuid)
	//assert.Equal(httpResponse.StatusCode, 200)
}

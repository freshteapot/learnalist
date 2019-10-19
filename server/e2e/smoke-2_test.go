package e2e_test

import (
	"fmt"
	"testing"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/e2e"
	"github.com/stretchr/testify/assert"
)

func TestSharePublic2(t *testing.T) {
	var messageResponse e2e.MessageResponse

	assert := assert.New(t)
	username := generateUsername()
	learnalistClient := e2e.NewClient(server)
	fmt.Printf("> Create user %s\n", username)
	userInfoOwner := learnalistClient.Register(username, password)

	listInfo, _ := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
	messageResponse = learnalistClient.SetListShare(userInfoOwner, listInfo.Uuid, "public")
	assert.Equal(messageResponse.Message, "List is now public")
	fmt.Println(fmt.Sprintf("http://localhost:1234/alists/%s.html", listInfo.Uuid))
}

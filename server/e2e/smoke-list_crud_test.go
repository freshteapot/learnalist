package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/e2e"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/stretchr/testify/assert"
)

func TestListCrud(t *testing.T) {
	assert := assert.New(t)
	learnalistClient := e2e.NewClient(server)
	userInfoOwner := learnalistClient.Register(usernameOwner, password)
	aList, _ := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(""))
	assert.NotEmpty(aList.Uuid)
	b, _ := json.Marshal(aList)
	aListB, _ := learnalistClient.PutListV1(userInfoOwner, aList.Uuid, string(b))
	assert.Equal(aList.Uuid, aListB.Uuid)
	resp, _ := learnalistClient.RawPutListV1(userInfoOwner, "aa", string(b))
	assert.Equal(resp.StatusCode, http.StatusBadRequest)
}

func TestListSharing(t *testing.T) {
	var aList alist.Alist
	var aListB alist.Alist
	var input []byte

	assert := assert.New(t)
	learnalistClient := e2e.NewClient(server)
	userInfoOwner := learnalistClient.Register(usernameOwner, password)
	fmt.Println("> By default shared privately")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(""))
	assert.Equal(aList.Info.SharedWith, aclKeys.NotShared)
	fmt.Println("> Explicitly share privately")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(aclKeys.NotShared))
	assert.Equal(aList.Info.SharedWith, aclKeys.NotShared)
	fmt.Println("> Explicitly share with public")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(aclKeys.SharedWithPublic))
	assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithPublic)
	fmt.Println("Explicitly share with friends")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(aclKeys.SharedWithFriends))
	assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithFriends)

	fmt.Println("> Share privately and then set it to public")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(aclKeys.NotShared))
	assert.Equal(aList.Info.SharedWith, aclKeys.NotShared)
	aList.Info.SharedWith = aclKeys.SharedWithPublic
	input, _ = json.Marshal(aList)
	aListB, _ = learnalistClient.PutListV1(userInfoOwner, aList.Uuid, string(input))
	assert.Equal(aListB.Info.SharedWith, aclKeys.SharedWithPublic)
	assert.Equal(aList, aListB)

	fmt.Println("> Share publicly first, then set it to private")
	aListB = alist.Alist{}
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(aclKeys.SharedWithPublic))
	assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithPublic)
	aList.Info.SharedWith = aclKeys.NotShared
	input, _ = json.Marshal(aList)
	aListB, _ = learnalistClient.PutListV1(userInfoOwner, aList.Uuid, string(input))
	assert.Equal(aListB.Info.SharedWith, aclKeys.NotShared)
	assert.Equal(aList, aListB)

	fmt.Println("> Share publicly first, then set it to friends")
	aListB = alist.Alist{}
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(aclKeys.SharedWithPublic))
	assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithPublic)
	aList.Info.SharedWith = aclKeys.SharedWithFriends
	input, _ = json.Marshal(aList)
	aListB, _ = learnalistClient.PutListV1(userInfoOwner, aList.Uuid, string(input))
	assert.Equal(aListB.Info.SharedWith, aclKeys.SharedWithFriends)
	assert.Equal(aList, aListB)
}

func TestLabelCrud(t *testing.T) {
	assert := assert.New(t)
	learnalistClient := e2e.NewClient(server)

	userInfoOwner := learnalistClient.Register(usernameOwner, password)
	resp, err := learnalistClient.RawPostLabelV1(userInfoOwner, "water")
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusCreated)
	labels, err := learnalistClient.PostLabelV1(userInfoOwner, "fire")
	assert.NoError(err)
	assert.Equal(labels, []string{"fire", "water"})

	resp, err = learnalistClient.RawDeleteLabelV1(userInfoOwner, "water")
	assert.NoError(err)
	labels, err = learnalistClient.PostLabelV1(userInfoOwner, "fire")
	assert.NoError(err)
	assert.Equal(labels, []string{"fire"})
}

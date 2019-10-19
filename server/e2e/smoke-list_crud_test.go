package e2e_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	aList, _ := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
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
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
	assert.Equal(aList.Info.SharedWith, aclKeys.NotShared)
	fmt.Println("> Explicitly share privately")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.NotShared))
	assert.Equal(aList.Info.SharedWith, aclKeys.NotShared)
	fmt.Println("> Explicitly share with public")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.SharedWithPublic))
	assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithPublic)
	fmt.Println("Explicitly share with friends")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.SharedWithFriends))
	assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithFriends)

	fmt.Println("> Share privately and then set it to public")
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.NotShared))
	assert.Equal(aList.Info.SharedWith, aclKeys.NotShared)
	aList.Info.SharedWith = aclKeys.SharedWithPublic
	input, _ = json.Marshal(aList)
	aListB, _ = learnalistClient.PutListV1(userInfoOwner, aList.Uuid, string(input))
	assert.Equal(aListB.Info.SharedWith, aclKeys.SharedWithPublic)
	assert.Equal(aList, aListB)

	fmt.Println("> Share publicly first, then set it to private")
	aListB = alist.Alist{}
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.SharedWithPublic))
	assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithPublic)
	aList.Info.SharedWith = aclKeys.NotShared
	input, _ = json.Marshal(aList)
	aListB, _ = learnalistClient.PutListV1(userInfoOwner, aList.Uuid, string(input))
	assert.Equal(aListB.Info.SharedWith, aclKeys.NotShared)
	assert.Equal(aList, aListB)

	fmt.Println("> Share publicly first, then set it to friends")
	aListB = alist.Alist{}
	aList, _ = learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.SharedWithPublic))
	assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithPublic)
	aList.Info.SharedWith = aclKeys.SharedWithFriends
	input, _ = json.Marshal(aList)
	aListB, _ = learnalistClient.PutListV1(userInfoOwner, aList.Uuid, string(input))
	assert.Equal(aListB.Info.SharedWith, aclKeys.SharedWithFriends)
	assert.Equal(aList, aListB)
}

func TestLabelCrud(t *testing.T) {
	var labels []string
	var err error
	var resp *http.Response

	assert := assert.New(t)
	learnalistClient := e2e.NewClient(server)
	fmt.Println("> Register user")
	userInfoOwner := learnalistClient.Register(usernameOwner, password)

	fmt.Println("> Empty list of labels")
	labels, err = learnalistClient.GetLabelsByMeV1(userInfoOwner)
	assert.NoError(err)
	assert.Equal(labels, []string{})

	fmt.Println("> Add water as a label")
	resp, err = learnalistClient.RawPostLabelV1(userInfoOwner, "water")
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusCreated)

	fmt.Println("> Add fire as a label")
	labels, err = learnalistClient.PostLabelV1(userInfoOwner, "fire")
	assert.NoError(err)
	assert.Equal(labels, []string{"fire", "water"})

	fmt.Println("> Make sure adding fire twice, results in only fire appearing once")
	labels, err = learnalistClient.PostLabelV1(userInfoOwner, "fire")
	assert.NoError(err)
	assert.Equal(labels, []string{"fire", "water"})
	labels, err = learnalistClient.GetLabelsByMeV1(userInfoOwner)
	assert.NoError(err)
	assert.Equal(labels, []string{"fire", "water"})

	fmt.Println("> Remove water")
	resp, err = learnalistClient.RawDeleteLabelV1(userInfoOwner, "water")
	assert.NoError(err)
	labels, err = learnalistClient.GetLabelsByMeV1(userInfoOwner)
	assert.NoError(err)
	assert.Equal(labels, []string{"fire"})

	fmt.Println("> Remove fire and make sure the user has no labels")
	resp, err = learnalistClient.RawDeleteLabelV1(userInfoOwner, "fire")
	labels, err = learnalistClient.GetLabelsByMeV1(userInfoOwner)
	assert.NoError(err)
	assert.Equal(labels, []string{})
}

func TestUserHasEmptyLists(t *testing.T) {
	assert := assert.New(t)
	username := generateUsername()
	learnalistClient := e2e.NewClient(server)
	fmt.Printf("> Create user %s\n", username)
	userInfoOwner := learnalistClient.Register(username, password)
	resp, err := learnalistClient.RawGetListsByMe(userInfoOwner)
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)

	fmt.Println("> Empty list")
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)
	assert.Equal(cleanEchoJSONResponse(data), `[]`)
}

func TestUserHasTwoLists(t *testing.T) {
	assert := assert.New(t)
	username := generateUsername()
	learnalistClient := e2e.NewClient(server)
	fmt.Printf("> Create user %s\n", username)
	userInfoOwner := learnalistClient.Register(username, password)
	learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
	learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.FromToList, ""))

	resp, err := learnalistClient.RawGetListsByMe(userInfoOwner)
	assert.NoError(err)
	assert.Equal(resp.StatusCode, http.StatusOK)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)
	fmt.Println(cleanEchoJSONResponse(data))
}

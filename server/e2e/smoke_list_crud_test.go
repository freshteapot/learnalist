package e2e_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/e2e"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("list crud", func() {

		assert := assert.New(GinkgoT())
		username := generateUsername()
		learnalistClient := e2e.NewClient(server)
		fmt.Printf("> Create user %s\n", username)
		userInfoOwner := learnalistClient.Register(username, password)

		aList, _ := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		assert.NotEmpty(aList.Uuid)
		b, _ := json.Marshal(aList)
		aListB, _ := learnalistClient.PutListV1(userInfoOwner, aList.Uuid, string(b))
		assert.Equal(aList.Uuid, aListB.Uuid)
		resp, _ := learnalistClient.RawPutListV1(userInfoOwner, "aa", string(b))
		assert.Equal(resp.StatusCode, http.StatusBadRequest)
	})
	It("list sharing", func() {

		var aList alist.Alist
		var aListB alist.Alist
		var input []byte

		assert := assert.New(GinkgoT())
		username := generateUsername()
		learnalistClient := e2e.NewClient(server)
		fmt.Printf("> Create user %s\n", username)
		userInfoOwner := learnalistClient.Register(username, password)

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
	})

	It("label crud", func() {

		var labels []string
		var err error
		var resp *http.Response

		assert := assert.New(GinkgoT())
		username := generateUsername()
		learnalistClient := e2e.NewClient(server)
		fmt.Printf("> Create user %s\n", username)
		userInfoOwner := learnalistClient.Register(username, password)

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
	})
	It("user has empty lists", func() {

		assert := assert.New(GinkgoT())
		username := generateUsername()
		learnalistClient := e2e.NewClient(server)
		fmt.Printf("> Create user %s\n", username)
		userInfoOwner := learnalistClient.Register(username, password)
		resp, err := learnalistClient.RawGetListsByMe(userInfoOwner, "", "")
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusOK)

		fmt.Println("> Empty list")
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(err)
		assert.Equal(testutils.CleanEchoResponseFromByte(data), `[]`)
	})
	It("user has two list v1 and v2", func() {

		assert := assert.New(GinkgoT())
		username := generateUsername()
		learnalistClient := e2e.NewClient(server)
		fmt.Printf("> Create user %s\n", username)
		userInfoOwner := learnalistClient.Register(username, password)
		learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.FromToList, ""))

		resp, err := learnalistClient.RawGetListsByMe(userInfoOwner, "", "")
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusOK)

		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(err)

		var aLists []*alist.Alist
		err = json.Unmarshal(data, &aLists)
		assert.NoError(err)
		assert.Equal(aLists[0].Info.ListType, alist.SimpleList)
		assert.Equal(aLists[1].Info.ListType, alist.FromToList)
	})
	It("alist v3", func() {

		assert := assert.New(GinkgoT())
		username := generateUsername()
		learnalistClient := e2e.NewClient(server)
		fmt.Printf("> Create user %s\n", username)
		userInfoOwner := learnalistClient.Register(username, password)
		aList, err := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.Concept2, ""))
		assert.NoError(err)
		assert.Equal(aList.Info.Labels, []string{"rowing", "concept2"})
		aList.Info.Labels = []string{}
		b, _ := json.Marshal(aList)
		aList2, err := learnalistClient.PutListV1(userInfoOwner, aList.Uuid, string(b))
		assert.NoError(err)
		assert.Equal(aList.Uuid, aList2.Uuid)
		assert.Equal(aList2.Info.Labels, []string{"rowing", "concept2"})
	})
	It("alist filter", func() {

		var aLists []*alist.Alist
		var err error
		assert := assert.New(GinkgoT())
		username := generateUsername()
		learnalistClient := e2e.NewClient(server)
		fmt.Printf("> Create user %s\n", username)
		userInfoOwner := learnalistClient.Register(username, password)
		lists := []string{`
{
	"data": ["car"],
	"info": {
		"title": "Days of the Week",
		"type": "v1"
	}
}
`,
			`
{
	"data": [],
	"info": {
		"title": "Days of the Week 2",
		"type": "v1"
	}
}
`,
			`{
	"data": ["car"],
	"info": {
		"title": "Days of the Week",
		"type": "v1",
		"labels": [
			"car",
			"water"
		]
	}
}`,
			`{
	"data": [{"from":"car", "to": "bil"}],
	"info": {
		"title": "Days of the Week",
		"type": "v2",
			"labels": [
			"water"
		]
	}
}`,
		}

		uuids := []string{}
		type uuidOnly struct {
			Uuid string `json:"uuid"`
		}

		for _, item := range lists {
			aList, err := learnalistClient.PostListV1(userInfoOwner, item)
			assert.NoError(err)
			uuids = append(uuids, aList.Uuid)
		}

		aLists, err = learnalistClient.GetListsByMe(userInfoOwner, "", "")
		assert.NoError(err)
		assert.Equal(4, len(aLists))

		aLists, err = learnalistClient.GetListsByMe(userInfoOwner, "water", "")
		assert.NoError(err)
		assert.Equal(2, len(aLists))

		aLists, err = learnalistClient.GetListsByMe(userInfoOwner, "", alist.SimpleList)
		assert.NoError(err)
		assert.Equal(3, len(aLists))

		aLists, err = learnalistClient.GetListsByMe(userInfoOwner, "", alist.FromToList)
		assert.NoError(err)
		assert.Equal(1, len(aLists))

		aLists, err = learnalistClient.GetListsByMe(userInfoOwner, "car,water", "")
		assert.NoError(err)
		assert.Equal(2, len(aLists))

		aLists, err = learnalistClient.GetListsByMe(userInfoOwner, "car,water", alist.FromToList)
		assert.NoError(err)
		assert.Equal(1, len(aLists))

		aLists, err = learnalistClient.GetListsByMe(userInfoOwner, "card", "")
		assert.NoError(err)
		assert.Equal(0, len(aLists))
	})

	It("method not supported for saving list", func() {
		username := generateUsername()
		learnalistClient := e2e.NewClient(server)
		fmt.Printf("> Create user %s\n", username)
		userInfoOwner := learnalistClient.Register(username, password)

		uri := "/api/v1/alist"
		resp, err := learnalistClient.RawV1(userInfoOwner, http.MethodDelete, uri, "")
		Expect(err).To(BeNil())
		Expect(resp.StatusCode, http.StatusMethodNotAllowed)
	})

	It("delete alist not found", func() {
		username := generateUsername()
		learnalistClient := e2e.NewClient(server)
		fmt.Printf("> Create user %s\n", username)
		userInfoOwner := learnalistClient.Register(username, password)

		alistUUID := "fake"
		resp, err := learnalistClient.RawDeleteListV1(userInfoOwner, alistUUID)
		defer resp.Body.Close()
		Expect(err).To(BeNil())
		Expect(resp.StatusCode, http.StatusNotFound)
		testutils.CheckMessageResponseFromReader(resp.Body, i18n.SuccessAlistNotFound)
	})

	It("only owner of the list can alter it", func() {

		var raw map[string]interface{}
		assert := assert.New(GinkgoT())
		learnalistClient := e2e.NewClient(server)
		userInfoOwner := learnalistClient.Register(usernameOwner, password)
		userInfoReader := learnalistClient.Register(usernameReader, password)

		aList, err := learnalistClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		assert.NoError(err)
		assert.NotEmpty(aList.Uuid)

		b, _ := json.Marshal(aList)
		resp, err := learnalistClient.RawPutListV1(userInfoOwner, aList.Uuid, string(b))
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusOK)

		resp, err = learnalistClient.RawPutListV1(userInfoReader, aList.Uuid, string(b))
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusForbidden)

		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(err)
		json.Unmarshal(data, &raw)
		assert.Equal(raw["message"].(string), i18n.InputSaveAlistOperationOwnerOnly)

		resp, err = learnalistClient.RawDeleteListV1(userInfoReader, aList.Uuid)
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusForbidden)
		defer resp.Body.Close()
		data, err = ioutil.ReadAll(resp.Body)
		assert.NoError(err)
		json.Unmarshal(data, &raw)
		assert.Equal(raw["message"].(string), i18n.InputDeleteAlistOperationOwnerOnly)
	})
})

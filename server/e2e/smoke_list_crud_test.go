package e2e_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Testing with Ginkgo", func() {
	It("list crud", func() {
		assert := assert.New(GinkgoT())
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)

		aList, _ := e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		assert.NotEmpty(aList.Uuid)
		b, _ := json.Marshal(aList)
		aListB, _ := e2eClient.PutListV1(userInfoOwner, aList.Uuid, string(b))
		assert.Equal(aList.Uuid, aListB.Uuid)
		resp, _ := e2eClient.RawPutListV1(userInfoOwner, "aa", string(b))
		assert.Equal(resp.StatusCode, http.StatusBadRequest)
		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
	})

	It("list sharing", func() {

		var (
			aList  alist.Alist
			aListB alist.Alist
			input  []byte
		)

		assert := assert.New(GinkgoT())
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)
		fmt.Println("> By default shared privately")
		aList, _ = e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		assert.Equal(aList.Info.SharedWith, aclKeys.NotShared)
		fmt.Println("> Explicitly share privately")
		aList, _ = e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.NotShared))
		assert.Equal(aList.Info.SharedWith, aclKeys.NotShared)
		fmt.Println("> Explicitly share with public")
		aList, _ = e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.SharedWithPublic))
		assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithPublic)
		fmt.Println("Explicitly share with friends")
		aList, _ = e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.SharedWithFriends))
		assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithFriends)

		fmt.Println("> Share privately and then set it to public")
		aList, _ = e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.NotShared))
		assert.Equal(aList.Info.SharedWith, aclKeys.NotShared)
		aList.Info.SharedWith = aclKeys.SharedWithPublic
		input, _ = json.Marshal(aList)
		aListB, _ = e2eClient.PutListV1(userInfoOwner, aList.Uuid, string(input))
		assert.Equal(aListB.Info.SharedWith, aclKeys.SharedWithPublic)
		assert.Equal(aList, aListB)

		fmt.Println("> Share publicly first, then set it to private")
		aListB = alist.Alist{}
		aList, _ = e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.SharedWithPublic))
		assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithPublic)
		aList.Info.SharedWith = aclKeys.NotShared
		input, _ = json.Marshal(aList)
		aListB, _ = e2eClient.PutListV1(userInfoOwner, aList.Uuid, string(input))
		assert.Equal(aListB.Info.SharedWith, aclKeys.NotShared)
		assert.Equal(aList, aListB)

		fmt.Println("> Share publicly first, then set it to friends")
		aListB = alist.Alist{}
		aList, _ = e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, aclKeys.SharedWithPublic))
		assert.Equal(aList.Info.SharedWith, aclKeys.SharedWithPublic)
		aList.Info.SharedWith = aclKeys.SharedWithFriends
		input, _ = json.Marshal(aList)
		aListB, _ = e2eClient.PutListV1(userInfoOwner, aList.Uuid, string(input))
		assert.Equal(aListB.Info.SharedWith, aclKeys.SharedWithFriends)
		assert.Equal(aList, aListB)

		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
	})

	It("label crud", func() {

		var labels []string
		var err error
		var resp *http.Response

		assert := assert.New(GinkgoT())
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)
		fmt.Println("> Empty list of labels")
		labels, err = e2eClient.GetLabelsByMeV1(userInfoOwner)
		assert.NoError(err)
		assert.Equal(labels, []string{})

		fmt.Println("> Add water as a label")
		resp, err = e2eClient.RawPostLabelV1(userInfoOwner, "water")
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusCreated)

		fmt.Println("> Add fire as a label")
		labels, err = e2eClient.PostLabelV1(userInfoOwner, "fire")
		assert.NoError(err)
		assert.Equal(labels, []string{"fire", "water"})

		fmt.Println("> Make sure adding fire twice, results in only fire appearing once")
		labels, err = e2eClient.PostLabelV1(userInfoOwner, "fire")
		assert.NoError(err)
		assert.Equal(labels, []string{"fire", "water"})
		labels, err = e2eClient.GetLabelsByMeV1(userInfoOwner)
		assert.NoError(err)
		assert.Equal(labels, []string{"fire", "water"})

		fmt.Println("> Remove water")
		resp, err = e2eClient.RawDeleteLabelV1(userInfoOwner, "water")
		assert.NoError(err)
		labels, err = e2eClient.GetLabelsByMeV1(userInfoOwner)
		assert.NoError(err)
		assert.Equal(labels, []string{"fire"})

		fmt.Println("> Remove fire and make sure the user has no labels")
		resp, err = e2eClient.RawDeleteLabelV1(userInfoOwner, "fire")
		labels, err = e2eClient.GetLabelsByMeV1(userInfoOwner)
		assert.NoError(err)
		assert.Equal(labels, []string{})

		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
	})
	It("user has empty lists", func() {

		assert := assert.New(GinkgoT())
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)
		resp, err := e2eClient.RawGetListsByMe(userInfoOwner, "", "")
		assert.NoError(err)
		assert.Equal(resp.StatusCode, http.StatusOK)

		fmt.Println("> Empty list")
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		assert.NoError(err)
		assert.Equal(testutils.CleanEchoResponseFromByte(data), `[]`)

		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
	})

	It("user has two list v1 and v2", func() {

		assert := assert.New(GinkgoT())
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)
		e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.FromToList, ""))

		resp, err := e2eClient.RawGetListsByMe(userInfoOwner, "", "")
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
		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
	})

	It("alist v3", func() {
		assert := assert.New(GinkgoT())
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)
		aList, err := e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.Concept2, ""))
		assert.NoError(err)
		assert.Equal(aList.Info.Labels, []string{"rowing", "concept2"})
		aList.Info.Labels = []string{}
		b, _ := json.Marshal(aList)
		aList2, err := e2eClient.PutListV1(userInfoOwner, aList.Uuid, string(b))
		assert.NoError(err)
		assert.Equal(aList.Uuid, aList2.Uuid)
		assert.Equal(aList2.Info.Labels, []string{"rowing", "concept2"})
		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
	})
	It("alist filter", func() {

		var aLists []*alist.Alist
		var err error
		assert := assert.New(GinkgoT())
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)
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
			aList, err := e2eClient.PostListV1(userInfoOwner, item)
			assert.NoError(err)
			uuids = append(uuids, aList.Uuid)
		}

		aLists, err = e2eClient.GetListsByMe(userInfoOwner, "", "")
		assert.NoError(err)
		assert.Equal(4, len(aLists))

		aLists, err = e2eClient.GetListsByMe(userInfoOwner, "water", "")
		assert.NoError(err)
		assert.Equal(2, len(aLists))

		aLists, err = e2eClient.GetListsByMe(userInfoOwner, "", alist.SimpleList)
		assert.NoError(err)
		assert.Equal(3, len(aLists))

		aLists, err = e2eClient.GetListsByMe(userInfoOwner, "", alist.FromToList)
		assert.NoError(err)
		assert.Equal(1, len(aLists))

		aLists, err = e2eClient.GetListsByMe(userInfoOwner, "car,water", "")
		assert.NoError(err)
		assert.Equal(2, len(aLists))

		aLists, err = e2eClient.GetListsByMe(userInfoOwner, "car,water", alist.FromToList)
		assert.NoError(err)
		assert.Equal(1, len(aLists))

		aLists, err = e2eClient.GetListsByMe(userInfoOwner, "card", "")
		assert.NoError(err)
		assert.Equal(0, len(aLists))

		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
	})

	It("method not supported for saving list", func() {
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)

		uri := "/api/v1/alist"
		resp, err := e2eClient.RawV1(userInfoOwner, http.MethodDelete, uri, "")
		Expect(err).To(BeNil())
		Expect(resp.StatusCode, http.StatusMethodNotAllowed)
		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
	})

	It("delete alist not found", func() {
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)

		alistUUID := "fake"
		resp, err := e2eClient.RawDeleteListV1(userInfoOwner, alistUUID)
		defer resp.Body.Close()
		Expect(err).To(BeNil())
		Expect(resp.StatusCode, http.StatusNotFound)
		testutils.CheckMessageResponseFromReader(resp.Body, i18n.SuccessAlistNotFound)
		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
	})

	It("only owner of the list can alter it", func() {
		authOwner, userInfoOwner := RegisterAndLogin(openapiClient.API)
		authReader, userInfoReader := RegisterAndLogin(openapiClient.API)

		aList, err := e2eClient.PostListV1(userInfoOwner, getInputListWithShare(alist.SimpleList, ""))
		Expect(err).To(BeNil())
		Expect(aList.Uuid).To(Not(BeEmpty()))

		b, _ := json.Marshal(aList)
		resp, err := e2eClient.RawPutListV1(userInfoOwner, aList.Uuid, string(b))
		Expect(err).To(BeNil())
		Expect(resp.StatusCode, http.StatusOK)

		resp, err = e2eClient.RawPutListV1(userInfoReader, aList.Uuid, string(b))
		defer resp.Body.Close()
		Expect(err).To(BeNil())
		Expect(resp.StatusCode, http.StatusForbidden)
		testutils.CheckMessageResponseFromReader(resp.Body, i18n.InputSaveAlistOperationOwnerOnly)

		resp, err = e2eClient.RawDeleteListV1(userInfoReader, aList.Uuid)
		defer resp.Body.Close()
		Expect(err).To(BeNil())
		Expect(resp.StatusCode, http.StatusForbidden)
		testutils.CheckMessageResponseFromReader(resp.Body, i18n.InputDeleteAlistOperationOwnerOnly)

		// Delete user
		openapiClient.DeleteUser(authOwner, userInfoOwner.UserUuid)
		openapiClient.DeleteUser(authReader, userInfoReader.UserUuid)
	})
})

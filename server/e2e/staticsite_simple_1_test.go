package e2e_test

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Static Site Simple flow", func() {
	It("Register, login and delete", func() {
		fmt.Println(`
		- Register user
		- Login
		- Create List
		- Confirm List exists for public viewing via static site
		- Delete user
		- Confirm list is removed in part of the delete
		- Confirm public list returns 404
		`)
		// Register and Login
		auth, loginInfo := RegisterAndLogin(openapiClient.API)
		// Create a list
		aList, _ := e2eClient.PostListV1(loginInfo, getInputListWithShare(alist.SimpleList, aclKeys.SharedWithPublic))
		Expect(aList.Info.SharedWith).To(Equal(aclKeys.SharedWithPublic))
		// Wait to give the static site chance to build
		time.Sleep(1000 * time.Millisecond)

		httpResponse, err := e2eClient.GetAlistHtml(loginInfo, aList.Uuid)
		Expect(err).To(BeNil())
		Expect(httpResponse.StatusCode).To(Equal(http.StatusOK))
		Expect(strings.Contains(string(httpResponse.Body), "<title>Days of the Week</title>")).To(BeTrue())

		// Verify list exists in static site
		// Delete user
		openapiClient.DeleteUser(auth, loginInfo.UserUuid)

		// New user to verify
		auth, loginInfo = RegisterAndLogin(openapiClient.API)
		// Confirm removed
		time.Sleep(1000 * time.Millisecond)

		httpResponse, err = e2eClient.GetAlistHtml(loginInfo, aList.Uuid)
		Expect(err).To(BeNil())
		Expect(httpResponse.StatusCode).To(Equal(http.StatusNotFound))

		// Delete user
		openapiClient.DeleteUser(auth, loginInfo.UserUuid)
	})

})

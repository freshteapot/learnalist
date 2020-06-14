package e2e_test

import (
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/api"
	"github.com/freshteapot/learnalist-api/server/e2e"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Smoke user", func() {
	var learnalistClient e2e.Client

	BeforeEach(func() {
		learnalistClient = e2e.NewClient(server)
	})

	It("Register, login and delete", func() {
		username := generateUsername()
		learnalistClient.Register(username, password)

		response, err := learnalistClient.RawLogin(username, password)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
		defer response.Body.Close()

		decoder := json.NewDecoder(response.Body)
		var userCredentials api.HTTPLoginResponse
		err = decoder.Decode(&userCredentials)
		Expect(err).ShouldNot(HaveOccurred())

		statusCode, httpResponse := learnalistClient.DeleteUser(userCredentials)
		Expect(statusCode).To(Equal(http.StatusOK))
		Expect(convertResponseToString(httpResponse)).To(Equal(`{"message":"User has been removed"}`))
	})
})

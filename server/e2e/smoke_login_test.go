package e2e_test

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/e2e"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SmokeLogin", func() {
	var learnalistClient e2e.Client

	BeforeEach(func() {
		learnalistClient = e2e.NewClient(server)
	})

	It("Empty input", func() {
		response, err := learnalistClient.RawLogin("", "")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusBadRequest))
	})

	It("Not a valid user", func() {
		username := generateUsername()
		learnalistClient.Register(username, password)

		response, err := learnalistClient.RawLogin(username, "hello123")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusForbidden))
	})

	It("Valid user", func() {
		username := generateUsername()
		learnalistClient.Register(username, password)

		response, err := learnalistClient.RawLogin(username, password)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(response.StatusCode).To(Equal(http.StatusOK))
	})
})

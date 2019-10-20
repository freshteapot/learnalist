package server_test

import (
	"github.com/freshteapot/learnalist-api/server/alists/server"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing server that handles static and html etc", func() {

	When("Requesting a list", func() {
		It("Valid html request", func() {
			input := "/alists/b453a069-24e4-5b52-8de5-23ac05c753ef.html"
			alistUUID, isA, err := server.GetAlistUUIDFromUrl(input)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(alistUUID).To(Equal("b453a069-24e4-5b52-8de5-23ac05c753ef"))
			Expect(isA).To(Equal("html"))
		})

		It("Valid json request", func() {
			input := "/alists/b453a069-24e4-5b52-8de5-23ac05c753ef.json"
			alistUUID, isA, err := server.GetAlistUUIDFromUrl(input)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(alistUUID).To(Equal("b453a069-24e4-5b52-8de5-23ac05c753ef"))
			Expect(isA).To(Equal("json"))
		})

		It("The uri has an extra / in it, making it invalid", func() {
			input := "/alists/b453a069-24e4-5b52-8de5-23ac05c753ef/a.json"
			_, _, err := server.GetAlistUUIDFromUrl(input)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("Invalid uri"))
		})

		It("The uri has an extra / in it, making it invalid", func() {
			input := "/alists/b453a069-24e4-5b52-8de5-23ac05c753efa.txt"
			_, _, err := server.GetAlistUUIDFromUrl(input)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).To(Equal("Unsupported format"))
		})
	})
})

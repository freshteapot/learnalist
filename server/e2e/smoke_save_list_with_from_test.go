package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Confirm logic around lists with from object in the info block", func() {
	When("Not present", func() {
		It("Should be normal behaviour", func() {
			Expect(true).To(BeTrue())
		})
	})

	When("from is present in the info block", func() {
		When("Saving a new list", func() {
			It("from kind is allowed", func() {

			})

			It("Make sure shared is set to private", func() {

			})

			It("from kind is not allowed", func() {

			})
		})

		When("Updating a list", func() {
			It("Make sure the from block cant be modified", func() {

			})

			When("from not learnalist", func() {
				It("Make sure shared is not modified", func() {

				})
			})

			When("from learnalist", func() {
				It("Shared can be changed", func() {

				})
			})
		})
	})
})

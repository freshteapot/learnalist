package user_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing User from IDP", func() {
	It("Bad json input", func() {
		Expect("TODO").To(Equal("TODO"))
	})

	It("Idp not enabled / supported", func() {

	})

	It("Defense code, if we add idp but do not add the logic", func() {

	})

	It("Failed to get userUUID from the idp", func() {

	})

	When("Looking up the user", func() {
		It("Issue talking to via the repo", func() {

		})

		When("User not found, register", func() {
			It("Issue talking to via the repo", func() {

			})

			It("New User registered", func() {

			})
		})

		It("Failed to create session", func() {

		})

		It("Session created", func() {

		})
	})
})

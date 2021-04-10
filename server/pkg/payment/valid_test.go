package payment_test

import (
	"github.com/freshteapot/learnalist-api/server/pkg/payment"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thoas/go-funk"
)

var _ = Describe("Testing Stuff and things", func() {
	It("Checking func.contains", func() {
		options := []payment.PaymentOption{
			{
				ID: "Test",
			},
		}

		find := "Test"
		want := funk.Contains(options, func(option payment.PaymentOption) bool {
			return option.ID == find
		})

		Expect(want).To(BeTrue())
	})
})

package remind_test

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Remind", func() {
	It("Validate time_of_day", func() {
		want := errors.New("fail")
		tests := []struct {
			input  string
			expect error
		}{
			{
				input:  "00:00:00",
				expect: want,
			},
			{
				input:  "car:00",
				expect: want,
			},
			{
				input:  "00:car",
				expect: want,
			},
			{
				input:  "00:00",
				expect: nil,
			},
			{
				input:  "000:00",
				expect: want,
			},
			{
				input:  "00:000",
				expect: want,
			},
			{
				input:  "25:00",
				expect: want,
			},
			{
				input:  "-1:60", // under 0
				expect: want,
			},
			{
				input:  "00:60", // under 0
				expect: want,
			},
			{
				input:  "00:-1", // under 0
				expect: want,
			},
		}

		for _, test := range tests {
			err := remind.ValidateTimeOfDay(test.input)
			if test.expect != nil {
				Expect(err).To(Equal(test.expect))
				continue
			}
			Expect(err).To(BeNil())
		}
	})

})

package remind_test

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing API", func() {

	It("", func() {
		want := errors.New("fail")
		tests := []struct {
			input  string
			expect error
		}{
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

		a := "00:00"
		parts := strings.Split(a, ":")
		fmt.Println(len(parts) == 2)

		hour, err := strconv.Atoi(parts[0])
		fmt.Println(err)
		minute, err := strconv.Atoi(parts[1])
		fmt.Println(err)

		fmt.Println(hour < 0)
		fmt.Println(hour > 23)
		fmt.Println(len(parts[0]) > 2)

		fmt.Println(minute < 0)
		fmt.Println(minute > 59)
		fmt.Println(len(parts[1]) > 2)

		Expect(remind.ValidateTimeOfDay(a)).To(BeNil())
	})

})

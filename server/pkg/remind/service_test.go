package remind_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TimeIn(t time.Time, name string) (time.Time, error) {

	loc, err := time.LoadLocation(name)
	if err == nil {
		t = t.In(loc)
	}
	return t, err
}

var _ = Describe("Testing Remind", func() {
	It("timezone fun", func() {
		/*
			- Load stream

			- Get list of timezones - from db
			- Loop over check map for time - from db
			- If found lookup users
			- Has record? yes OR no.
			- Send message
				- Will need token
				- Will need display_name
			- Send record back into topic, to make sure its still in the system
		*/
		for _, name := range []string{
			"",
			"Local",
			"Europe/Oslo",
			"Asia/Shanghai",
			"America/Metropolis",
		} {
			t, err := TimeIn(time.Now(), name)
			if err == nil {
				fmt.Println(t.Location(), t.Format("15:04"))
			} else {
				fmt.Println(name, "<time unknown>")
			}
		}
	})
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

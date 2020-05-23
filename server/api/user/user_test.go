package user_test

import (
	"github.com/freshteapot/learnalist-api/server/api/user"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing user login endpoint", func() {
	var (
		err         error
		cleanedUser user.RegisterInput
		input       user.RegisterInput
	)

	It("Test cases", func() {
		tests := []struct {
			Username    string
			Password    string
			Description string
			ExpectError bool
		}{
			{
				Username:    "",
				Password:    "",
				Description: "Fail because empty",
			},
			{
				Username:    "abc",
				Password:    "",
				Description: "Fail because username is too short and password is empty",
			},
			{
				Username:    "iamusera",
				Password:    "",
				Description: "Fail because password is empty",
			},
			{
				Username:    "iamusera",
				Password:    "test12",
				Description: "Fail because password is not long enough",
			},
			{
				Username:    "iamusera?",
				Password:    "test1234",
				Description: "Fail because username is not valid",
			},
		}

		for _, test := range tests {
			input = user.RegisterInput{
				Username: test.Username,
				Password: test.Password,
			}
		}
		_, err = user.Validate(input)
		Expect(err).NotTo(BeNil())
	})

	It("Confirm filtering of username to lowercase", func() {
		input.Username = "iamauserA"
		input.Password = "_i_am_a_test_"
		cleanedUser, err = user.Validate(input)
		Expect(cleanedUser.Username).To(Equal("iamausera"))
		Expect(err).To(BeNil())
	})
})

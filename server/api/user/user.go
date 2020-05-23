package user

import (
	"errors"

	"github.com/gookit/validate"
)

type RegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
}

/*
Username is required, needs to be at least 5 characters and can only be letters, numbers _ and -.
Password is required, minimum length of 7.
*/
func Validate(input RegisterInput) (RegisterInput, error) {
	var cleaned RegisterInput

	v := validate.New(&input)
	v.StopOnError = false
	v.StringRule("username", "required|minLen:5|alphaDash")
	v.FilterRule("username", "lower")
	v.StringRule("password", "required|minLen:7")

	v.Sanitize()
	if v.Validate() {
		v.BindSafeData(&cleaned)
		return cleaned, nil
	}
	return cleaned, errors.New("please refer to the documentation")
}

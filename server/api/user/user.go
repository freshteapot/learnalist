package user

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/gookit/validate"
)

/*
Username is required, needs to be at least 5 characters and can only be letters, numbers _ and -.
Password is required, minimum length of 7.
*/
func Validate(input api.HTTPUserRegisterInput) (api.HTTPUserRegisterInput, error) {
	var cleaned api.HTTPUserRegisterInput

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

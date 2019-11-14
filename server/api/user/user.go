package user

import (
	"errors"

	"github.com/gookit/validate"
)

type RegisterInput struct {
	Username string `json:"username" filter:"lower" validate:"required|minLen:5|alphaDash"`
	Password string `json:"password" validate:"required|minLen:7"`
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
	v := validate.New(input)
	v.StopOnError = false
	v.Sanitize()
	if v.Validate() {
		v.BindSafeData(&cleaned)
		return cleaned, nil
	}
	return cleaned, errors.New("please refer to the documentation")
}

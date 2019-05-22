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
	Uuid string `json:"uuid"`
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
	if v.Validate() { // validate ok
    v.BindSafeData(&cleaned)
		return cleaned, nil
	}
	// TODO write the documentation and have it installed as a list
	return cleaned, errors.New("Please refer to the documentation.")
	/*
		if input.Username == "" || input.Password == "" {
			return errors.New("Username and password should not be empty")
		}

		if strings.ToLower(input.Username) != input.Username {
			return errors.New("username needs to be lower case")
		}
		return nil
	*/
}

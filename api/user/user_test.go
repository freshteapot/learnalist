package user

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserSuite struct {
	suite.Suite
}

func (suite *UserSuite) SetupSuite() {

}

func (suite *UserSuite) SetupTest() {

}

func (suite *UserSuite) TearDownTest() {

}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}

func (suite *UserSuite) TestRegisterValidation() {
	var err error
	var cleanedUser RegisterInput
	input := &RegisterInput{
		Username: "",
		Password: "",
	}
	_, err = Validate(*input)
	suite.NotNil(err)
	input.Username = "abc"
	_, err = Validate(*input)
	suite.NotNil(err)
	input.Username = "iamauserA"
	input.Password = "_i_am_a_test_"
	cleanedUser, err = Validate(*input)
	suite.Nil(err)
	suite.Equal("iamausera", cleanedUser.Username)
}

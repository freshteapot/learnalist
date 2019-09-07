package authenticate_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthenticateSuite struct {
	suite.Suite
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(AuthenticateSuite))
}

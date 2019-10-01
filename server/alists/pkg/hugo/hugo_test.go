package hugo_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HugoSuite struct {
	suite.Suite
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(HugoSuite))
}

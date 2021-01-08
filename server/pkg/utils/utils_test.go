package utils_test

import (
	"testing"

	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/stretchr/testify/suite"
	"github.com/tj/assert"
)

type UtilsSuite struct {
	suite.Suite
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(UtilsSuite))
}

func TestStringArrayContains(t *testing.T) {
	var found bool
	items := []string{
		"morning",
		"evening",
		"day",
		"night",
	}
	found = utils.StringArrayContains(items, "snow")
	assert.False(t, found)

	found = utils.StringArrayContains(items, "morning")
	assert.True(t, found)
}

func TestStringArrayIndexOf(t *testing.T) {
	var found int
	items := []string{
		"morning",
		"evening",
		"day",
		"night",
	}
	found = utils.StringArrayIndexOf(items, "snow")
	assert.Equal(t, -1, found)

	found = utils.StringArrayIndexOf(items, "morning")
	assert.Equal(t, 0, found)

	found = utils.StringArrayIndexOf(items, "day")
	assert.Equal(t, 2, found)
}

func TestStringArrayRemoveAtIndex(t *testing.T) {
	items := []string{
		"morning",
		"evening",
		"day",
		"night",
	}
	index := 1
	after := utils.StringArrayRemoveAtIndex(items, index)
	assert.Equal(t, 3, len(after))
}

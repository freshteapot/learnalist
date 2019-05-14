package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringArrayContains(t *testing.T) {
	var found bool
	items := []string{
		"morning",
		"evening",
		"day",
		"night",
	}
	found = StringArrayContains(items, "snow")
	assert.False(t, found)

	found = StringArrayContains(items, "morning")
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
	found = StringArrayIndexOf(items, "snow")
	assert.Equal(t, -1, found)

	found = StringArrayIndexOf(items, "morning")
	assert.Equal(t, 0, found)

	found = StringArrayIndexOf(items, "day")
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
	after := StringArrayRemoveAtIndex(items, index)
	assert.Equal(t, 3, len(after))
}

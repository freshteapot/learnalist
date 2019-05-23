package alist

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTypeV1(t *testing.T) {
	aList := NewTypeV1()
	assert.Equal(t, SimpleList, aList.Info.ListType)
	assert.Equal(t, "TypeV1", reflect.TypeOf(aList.Data).Name())
}

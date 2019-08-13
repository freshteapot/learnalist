package alist

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTypeV2(t *testing.T) {
	aList := NewTypeV2()
	assert.Equal(t, FromToList, aList.Info.ListType)
	assert.Equal(t, "TypeV2", reflect.TypeOf(aList.Data).Name())
}

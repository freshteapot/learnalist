package alist

import (
	"fmt"
	"testing"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/stretchr/testify/assert"
)

var validAListTypeV1 = `{"data":["a","b"],"info":{"title":"I am a list","type":"v1"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`
var validAlistTypeV2 = `{"data":{"from":"car","to":"bil"},"info":{"title":"I am a list with items","type":"v2"},"uuid":"efeb4a6e-9a03-5aff-b46d-7f2ba1d7e7f9"}`

func TestValidateAlistInfo(t *testing.T) {
	jsonBytes := []byte(validAListTypeV1)
	aList := new(Alist)
	aList.UnmarshalJSON(jsonBytes)
	assert.NoError(t, validateAListInfo(aList.Info))

	// Confirm it handles empty title as we want.
	aList.Info.Title = ""
	err := validateAListInfo(aList.Info)
	assert.Equal(t, err.Error(), "Title cannot be empty.")
}

func TestAlistTypeV1(t *testing.T) {
	var err error
	var items AlistTypeV1
	jsonBytes := []byte(validAListTypeV1)
	aList := new(Alist)
	aList.UnmarshalJSON(jsonBytes)
	// This is valid with data
	err = validateAlistTypeV1(*aList)
	assert.Equal(t, err, nil)

	// This is not valid as it has an empty record
	items = AlistTypeV1{""}
	aList.Data = items

	err = validateAlistTypeV1(*aList)
	assert.Equal(t, err.Error(), "Item cant be empty at position 0")

	items = AlistTypeV1{"", ""}
	aList.Data = items

	err = validateAlistTypeV1(*aList)
	assert.Equal(t, err.Error(), "Item cant be empty at position 0\nItem cant be empty at position 1")

	items = AlistTypeV1{"a", "", "c"}
	aList.Data = items

	err = validateAlistTypeV1(*aList)
	assert.Equal(t, err.Error(), "Item cant be empty at position 1")
}

func TestAlistTypeV2(t *testing.T) {
	var err error
	var items AlistTypeV2
	jsonBytes := []byte(validAlistTypeV2)
	aList := new(Alist)
	aList.UnmarshalJSON(jsonBytes)
	// This is valid with data
	err = validateAlistTypeV2(*aList)
	assert.Equal(t, err, nil)

	// This is not valid as it has an empty record
	items = AlistTypeV2{
		AlistItemTypeV2{
			From: "",
			To:   "",
		},
	}
	aList.Data = items

	err = validateAlistTypeV2(*aList)
	assert.Equal(t, err.Error(), "Item cant be empty at position 0")

	items = AlistTypeV2{
		AlistItemTypeV2{
			From: "car",
			To:   "bil",
		},
		AlistItemTypeV2{
			From: "",
			To:   "",
		},
		AlistItemTypeV2{
			From: "water",
			To:   "vann",
		},
	}
	aList.Data = items

	err = validateAlistTypeV2(*aList)
	assert.Equal(t, err.Error(), "Item cant be empty at position 1")
}

func TestValidateAlist(t *testing.T) {
	var err error
	jsonBytes := []byte(validAListTypeV1)
	aList := new(Alist)
	aList.UnmarshalJSON(jsonBytes)
	assert.NoError(t, Validate(*aList))

	aList.Info = AlistInfo{}
	err = Validate(*aList)
	assert.Equal(t, err.Error(), fmt.Sprintf(i18n.ValidationErrorList, "Title cannot be empty."))

	// We check the failed path, as we have specific tests for each lists validation.
	aList.Info = AlistInfo{Title: "I am a title", ListType: "v1"}
	aList.Data = AlistTypeV1{""}
	err = Validate(*aList)
	assert.Equal(t, err.Error(), "Failed to pass list type v1. Item cant be empty at position 0")

	aList.Info = AlistInfo{Title: "I am a title", ListType: "v2"}
	aList.Data = AlistTypeV2{AlistItemTypeV2{From: "", To: ""}}
	err = Validate(*aList)
	assert.Equal(t, err.Error(), "Failed to pass list type v2. Item cant be empty at position 0")

	aList.Info = AlistInfo{Title: "I am a title", ListType: "v3"}
	aList.Data = TypeV3{
		TypeV3Item{
			When: "",
			Overall: V3Split{
				Time:     "3:00.0",
				Spm:      15,
				Distance: 1000,
				P500:     "1:00.1",
			},
			Splits: []V3Split{},
		},
	}
	err = Validate(*aList)
	assert.Equal(t, err.Error(), "Failed to pass list type v3. When should be YYYY-MM-DD.")

	// Make sure we handle Unsupported lists
	aList.Info = AlistInfo{Title: "I am a title", ListType: "na"}
	aList.Data = nil
	err = Validate(*aList)
	assert.Equal(t, err.Error(), "Unsupported list type.")

	// Validate labels
	aList.Info = AlistInfo{
		Title:    "I am a title",
		ListType: "v1",
		Labels: []string{
			"",
		}}
	aList.Data = AlistTypeV1{""}
	err = Validate(*aList)
	assert.Equal(t, err.Error(), "Failed to pass list info. Label can not be empty at position 0")
	aList.Info.Labels[0] = "iam a long label and should go over the allowed limit"
	err = Validate(*aList)
	assert.Equal(t, err.Error(), "Failed to pass list info. Label must be 20 or less characters long at position 0")
}

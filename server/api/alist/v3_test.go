package alist

import (
	"reflect"
	"testing"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/stretchr/testify/assert"
)

func TestNewTypeV3(t *testing.T) {
	aList := NewTypeV3()
	assert.Equal(t, Concept2, aList.Info.ListType)
	assert.Equal(t, "TypeV3", reflect.TypeOf(aList.Data).Name())
	assert.Equal(t, 2, len(aList.Info.Labels))
	assert.Equal(t, "rowing", aList.Info.Labels[0])
	assert.Equal(t, "concept2", aList.Info.Labels[1])
}

func TestAlistTypeV3(t *testing.T) {
	input := `
{
  "info": {
      "title": "Getting my row on.",
      "type": "v3"
  },
  "data": [{
    "when": "2019-05-06",
    "overall": {
      "time": "7:15.9",
      "distance": 2000,
      "spm": 28,
      "p500": "1:48.9"
    },
    "splits": [
      {
        "time": "1:46.4",
        "distance": 500,
        "spm": 29,
        "p500": "1:58.0"
      }
    ]
  }]
}
`
	jsonBytes := []byte(input)
	aList := new(Alist)
	err := aList.UnmarshalJSON(jsonBytes)
	assert.Nil(t, err)
	assert.Equal(t, "2019-05-06", aList.Data.(TypeV3)[0].When)
	// Confirm the enrichment happened.
	assert.Equal(t, 2, len(aList.Info.Labels))

	err = validateTypeV3(*aList)
	assert.Nil(t, err)

	typeV3Item := aList.Data.(TypeV3)[0]
	typeV3Item.When = ""
	aList.Data.(TypeV3)[0] = typeV3Item
	err = validateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.When = "2019-05-06"

	typeV3Item.Overall.Distance = 0
	aList.Data.(TypeV3)[0] = typeV3Item
	err = validateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Overall.Distance = 2000

	typeV3Item.Overall.Spm = 9
	aList.Data.(TypeV3)[0] = typeV3Item
	err = validateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Overall.Spm = 28

	typeV3Item.Overall.Time = "1.0"
	aList.Data.(TypeV3)[0] = typeV3Item
	err = validateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Overall.Time = "7:15.9"

	typeV3Item.Overall.P500 = "1.0"
	aList.Data.(TypeV3)[0] = typeV3Item
	err = validateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Overall.P500 = "1:10.0"

	after := enrichTypeV3(*aList)
	assert.Equal(t, 2, len(after.Info.Labels))
	// Make sure we dont duplicate the labels
	after = enrichTypeV3(after)
	assert.Equal(t, 2, len(after.Info.Labels))

	// Test a bad split
	typeV3Item.Splits[0].Time = "1.0"
	aList.Data.(TypeV3)[0] = typeV3Item
	err = validateTypeV3(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV3, err.Error())
	typeV3Item.Splits[0].Time = "1:0.0"
}

func TestTypeV3(t *testing.T) {
	input := `[{
    "when": "2019-05-06",
    "overall": {
      "time": "7:15.9",
      "distance": 2000,
      "spm": 28,
      "p500": "1:48.9"
    },
    "splits": [
      {
        "time": "1.46.4",
        "distance": 500,
        "spm": 29,
        "p500": "1:58.0"
      }
    ]
  }]
`
	jsonBytes := []byte(input)
	_, err := parseTypeV3(jsonBytes)
	assert.Nil(t, err)
}

func TestValidateTypeV3Distance(t *testing.T) {
	var err error
	a := 2000
	err = validateTypeV3Distance(a)
	assert.Nil(t, err)
	b := 0
	err = validateTypeV3Distance(b)
	assert.Equal(t, "Distance should not be empty.", err.Error())
}

func TestValidateTypeV3When(t *testing.T) {
	var err error
	a := "2019-05-15"
	err = validateTypeV3When(a)
	assert.Nil(t, err)
	b := "2019-05-5"
	err = validateTypeV3When(b)
	assert.Equal(t, "When should be YYYY-MM-DD.", err.Error())
	c := "2019/05/01"
	err = validateTypeV3When(c)
	assert.Equal(t, "When should be YYYY-MM-DD.", err.Error())

	d := ""
	err = validateTypeV3When(d)
	assert.Equal(t, "When should be YYYY-MM-DD.", err.Error())
}

func TestValidateTypeV3Time(t *testing.T) {
	var err error
	a := "1:49.9"
	err = validateTypeV3Time(a)
	assert.Nil(t, err)
	b := "49.9"
	err = validateTypeV3Time(b)
	assert.Equal(t, "Time is not valid format.", err.Error())
	c := "1.0.1"
	err = validateTypeV3Time(c)
	assert.Equal(t, "Time is not valid format.", err.Error())

	d := ""
	err = validateTypeV3Time(d)
	assert.Equal(t, "Time should not be empty.", err.Error())

	e := "1:00:0"
	err = validateTypeV3Time(e)
	assert.Equal(t, "Time is not valid format.", err.Error())
}

func TestValidateTypeV3Spm(t *testing.T) {
	var err error
	a := 29
	err = validateTypeV3Spm(a)
	assert.Nil(t, err)
	b := 9
	err = validateTypeV3Spm(b)
	assert.Equal(t, "Stroke per minute should be between the range 10 and 50.", err.Error())
	c := 51
	err = validateTypeV3Spm(c)
	assert.Equal(t, "Stroke per minute should be between the range 10 and 50.", err.Error())
}

func TestValidateTypeV3P500(t *testing.T) {
	var err error
	a := "1:49.9"
	err = validateTypeV3P500(a)
	assert.Nil(t, err)
	b := "49.9"
	err = validateTypeV3P500(b)
	assert.Equal(t, "Per 500 is not valid format.", err.Error())
	c := "1.0.1"
	err = validateTypeV3P500(c)
	assert.Equal(t, "Per 500 is not valid format.", err.Error())

	d := ""
	err = validateTypeV3P500(d)
	assert.Equal(t, "Per 500 should not be empty.", err.Error())

	e := "1:00:0"
	err = validateTypeV3P500(e)
	assert.Equal(t, "Per 500 is not valid format.", err.Error())
}
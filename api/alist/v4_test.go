package alist

import (
	"reflect"
	"testing"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/stretchr/testify/assert"
)

var validAlistTypeV4 = `
{
  "info": {
      "title": "A list of content with urls.",
      "type": "v4"
  },
  "data": [{
		"content": "A fabulous quote about nothing.",
		"url": "https://learnalist.net/not-real"
  }]
}
`

func TestNewTypeV4(t *testing.T) {
	aList := NewTypeV4()
	assert.Equal(t, ContentAndUrl, aList.Info.ListType)
	assert.Equal(t, "TypeV4", reflect.TypeOf(aList.Data).Name())
}

func TestAlistTypeV4(t *testing.T) {
	content := "A fabulous quote about nothing."
	url := "https://learnalist.net/not-real"
	jsonBytes := []byte(validAlistTypeV4)
	aList := new(Alist)
	err := aList.UnmarshalJSON(jsonBytes)
	assert.Nil(t, err)
	assert.Equal(t, content, aList.Data.(TypeV4)[0].Content)
	err = Validate(*aList)
	assert.Nil(t, err)

	err = validateTypeV4(*aList)
	assert.Nil(t, err)

	typeV4Item := aList.Data.(TypeV4)[0]
	typeV4Item.Content = ""
	aList.Data.(TypeV4)[0] = typeV4Item
	err = validateTypeV4(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV4, err.Error())
	err = Validate(*aList)
	assert.Equal(t, "Failed to pass list type v4. Please refer to the documentation on list type v4", err.Error())
	typeV4Item.Content = content

	typeV4Item.Url = ""
	aList.Data.(TypeV4)[0] = typeV4Item
	err = validateTypeV4(*aList)
	assert.Equal(t, i18n.ValidationAlistTypeV4, err.Error())
	err = Validate(*aList)
	assert.Equal(t, "Failed to pass list type v4. Please refer to the documentation on list type v4", err.Error())
	typeV4Item.Url = url
}

package alist

import (
	"encoding/json"
	"testing"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalJSON(t *testing.T) {
	var err error
	var jsonBytes []byte
	badRawJson := `{a}`
	missingInfoJSON := `{"data": []}`
	badInfo := `{"info": ""}`
	missingDataJSON := `{"info": {"title": "I am a title"}}`

	jsonBytes = []byte(badRawJson)

	aList := new(Alist)
	err = aList.UnmarshalJSON(jsonBytes)
	assert.Equal(t, err.Error(), "Failed to parse list.")

	jsonBytes = []byte(missingInfoJSON)
	err = aList.UnmarshalJSON(jsonBytes)
	assert.Equal(t, err.Error(), "Failed to pass list. Info is missing.")

	jsonBytes = []byte(badInfo)
	err = aList.UnmarshalJSON(jsonBytes)
	assert.Equal(t, err.Error(), "Failed to pass list. Something wrong with info object.")

	jsonBytes = []byte(missingDataJSON)
	err = aList.UnmarshalJSON(jsonBytes)
	assert.Equal(t, err.Error(), "Failed to pass list. Data is missing.")
}

func TestUnmarshalJSONBadParseV1(t *testing.T) {
	var err error
	var jsonBytes []byte
	var jsonStr = `{"data":[{}],"info":{"title":"I am a list","type":"v1"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`

	jsonBytes = []byte(jsonStr)

	aList := new(Alist)
	err = aList.UnmarshalJSON(jsonBytes)
	assert.Equal(t, err.Error(), i18n.ValidationErrorListV1)
}

func TestUnmarshalJSONBadParseV2(t *testing.T) {
	var err error
	var jsonBytes []byte
	var jsonStr = `{"data":[""],"info":{"title":"I am a list","type":"v2"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`

	jsonBytes = []byte(jsonStr)

	aList := new(Alist)
	err = aList.UnmarshalJSON(jsonBytes)
	assert.Equal(t, err.Error(), i18n.ValidationErrorListV2)
}

func TestUnmarshalJSONBadParseV3(t *testing.T) {
	var err error
	var jsonBytes []byte
	var jsonStr = `{"data":[""],"info":{"title":"I am a list","type":"v3"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`

	jsonBytes = []byte(jsonStr)

	aList := new(Alist)
	err = aList.UnmarshalJSON(jsonBytes)
	assert.Equal(t, err.Error(), i18n.ValidationErrorListV3)
}

func TestUnmarshalJSONBadParseV4(t *testing.T) {
	var err error
	var jsonBytes []byte
	var jsonStr = `{"data":[""],"info":{"title":"I am a list","type":"v4"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`

	jsonBytes = []byte(jsonStr)

	aList := new(Alist)
	err = aList.UnmarshalJSON(jsonBytes)
	assert.Equal(t, err.Error(), i18n.ValidationErrorListV4)
}

func TestUnmarshalJSONUnsupportedListType(t *testing.T) {
	var err error
	var jsonBytes []byte
	var jsonStr = `{"data":[],"info":{"title":"I am a list","type":"na"},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`

	jsonBytes = []byte(jsonStr)

	aList := new(Alist)
	err = aList.UnmarshalJSON(jsonBytes)
	assert.Equal(t, err.Error(), "Unsupported list type.")
}

func TestMarshalJSON(t *testing.T) {
	var jsonBytes []byte
	var jsonStr = `{"data":[],"info":{"title":"I am a list","type":"v1","labels":[]},"uuid":"230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"}`

	jsonBytes = []byte(jsonStr)
	aList := new(Alist)
	assert.NoError(t, aList.UnmarshalJSON(jsonBytes))

	jsonBytes, _ = json.Marshal(aList)

	assert.Equal(t, jsonStr, string(jsonBytes))
}

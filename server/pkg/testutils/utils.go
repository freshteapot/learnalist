package testutils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// https://dave.cheney.net/2016/05/10/test-fixtures-in-go
func GetTestData(path string) []byte {
	data, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s", path))

	if err != nil {
		panic(err)
	}

	return data
}

// GetTestDataAsJSONOneline Convert json to oneline
func GetTestDataAsJSONOneline(path string) string {
	b := GetTestData(path)
	var obj interface{}
	err := json.Unmarshal(b, &obj)
	if err != nil {
		panic(err)
	}

	b, err = json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	return string(b)
}

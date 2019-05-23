package alist

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// TypeV1 list type v1
type TypeV1 []string

func NewTypeV1() *Alist {
	aList := &Alist{}

	aList.Info.ListType = SimpleList
	data := make(TypeV1, 0)
	aList.Data = data

	labels := make([]string, 0)
	aList.Info.Labels = labels

	return aList
}

func validateTypeV1(aList Alist) error {
	var err error
	var feedbackMessage string
	var feedback []string = []string{}

	items := aList.Data.(TypeV1)
	for index, item := range items {
		if item == "" {
			feedback = append(feedback, fmt.Sprintf("Item cant be empty at position %d", index))
		}
	}

	if len(feedback) != 0 {
		feedbackMessage = strings.Join(feedback, "\n")
		err = errors.New(feedbackMessage)
	}

	return err
}

func parseTypeV1(jsonBytes []byte) (TypeV1, error) {
	listData := new(TypeV1)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}

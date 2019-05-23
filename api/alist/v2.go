package alist

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// TypeV2Item Item in  TypeV2
type TypeV2Item struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type TypeV2 []TypeV2Item

func NewTypeV2() *Alist {
	aList := &Alist{}

	aList.Info.ListType = FromToList
	data := make(TypeV2, 0)
	aList.Data = data

	labels := make([]string, 0)
	aList.Info.Labels = labels

	return aList
}

func parseTypeV2(jsonBytes []byte) (TypeV2, error) {
	listData := new(TypeV2)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}

func validateTypeV2(aList Alist) error {
	var err error
	var feedbackMessage string
	var feedback []string = []string{}

	items := aList.Data.(TypeV2)
	for index, item := range items {
		if item.From == "" && item.To == "" {
			feedback = append(feedback, fmt.Sprintf("Item cant be empty at position %d", index))
		}
	}

	if len(feedback) != 0 {
		feedbackMessage = strings.Join(feedback, "\n")
		err = errors.New(feedbackMessage)
	}

	return err
}

package alist

import (
	"errors"
	"fmt"
	"strings"
)

func NewTypeV2() *Alist {
	aList := &Alist{}

	aList.Info.ListType = FromToList
	data := make(AlistTypeV2, 0)
	aList.Data = data

	labels := make([]string, 0)
	aList.Info.Labels = labels

	return aList
}

func validateTypeV2(aList Alist) error {
	var err error
	var feedbackMessage string
	var feedback []string = []string{}

	items := aList.Data.(AlistTypeV2)
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

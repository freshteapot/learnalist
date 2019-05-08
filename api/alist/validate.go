package alist

import (
	"errors"
	"fmt"
	"strings"
)

func Validate(aList Alist) error {
	var err error

	err = validateAListInfo(aList.Info)
	if err != nil {
		err = errors.New(fmt.Sprintf("Failed to pass list info. %s", err.Error()))
		return err
	}
	switch aList.Info.ListType {
	case SimpleList:
		err = validateAlistTypeV1(aList)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to pass list type v1. %s", err.Error()))
			return err
		}
	case FromToList:
		err = validateAlistTypeV2(aList)
		if err != nil {
			err = errors.New(fmt.Sprintf("Failed to pass list type v2. %s", err.Error()))
			return err
		}
	default:
		err = errors.New("Unsupported list type.")
		return err
	}
	return nil
}

func validateAListInfo(info AlistInfo) error {
	var err error
	var feedbackMessage string
	var feedback []string = []string{}

	if info.Title == "" {
		feedback = append(feedback, "Title cannot be empty.")
	}

	for index, item := range info.Labels {
		if item == "" {
			feedback = append(feedback, fmt.Sprintf("Label can not be empty at position %d", index))
		}
		if len(item) > 20 {
			feedback = append(feedback, fmt.Sprintf("Label must be 20 or less characters long at position %d", index))
		}
	}

	if len(feedback) != 0 {
		feedbackMessage = strings.Join(feedback, "\n")
		err = errors.New(feedbackMessage)
	}

	return err
}

func validateAlistTypeV1(aList Alist) error {
	var err error
	var feedbackMessage string
	var feedback []string = []string{}

	items := aList.Data.(AlistTypeV1)
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

func validateAlistTypeV2(aList Alist) error {
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

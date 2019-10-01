package alist

import (
	"errors"
	"fmt"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/utils"
)

func Validate(aList Alist) error {
	var err error

	err = validateAListInfo(aList.Info)
	if err != nil {
		err = errors.New(fmt.Sprintf("Failed to pass list info. %s", err.Error()))
		return err
	}

	if !utils.StringArrayContains(allowedListTypes, aList.Info.ListType) {
		err = errors.New("Unsupported list type.")
		return err
	}

	switch aList.Info.ListType {
	case SimpleList:
		err = validateTypeV1(aList)
	case FromToList:
		err = validateTypeV2(aList)
	case Concept2:
		err = validateTypeV3(aList)
	case ContentAndUrl:
		err = validateTypeV4(aList)
	}

	if err != nil {
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

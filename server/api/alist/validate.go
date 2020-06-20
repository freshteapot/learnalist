package alist

import (
	"errors"
	"fmt"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/utils"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
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

	var mapper AlistTypeMarshalJSON
	switch aList.Info.ListType {
	case SimpleList:
		mapper = NewMapToV1()
	case FromToList:
		mapper = NewMapToV2()
		break
	case Concept2:
		mapper = NewMapToV3()
		break
	case ContentAndUrl:
		mapper = NewMapToV4()
		break
	}
	return mapper.Validate(aList)
}

func validateAListInfo(info AlistInfo) error {
	var err error
	var feedbackMessage string
	var feedback []string = []string{}

	if info.Title == "" {
		feedback = append(feedback, "Title cannot be empty.")
	}

	switch info.SharedWith {
	case aclKeys.NotShared:
	case aclKeys.SharedWithPublic:
	case aclKeys.SharedWithFriends:
		break
	default:
		feedback = append(feedback, "Invalid option for info.shared_with")
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

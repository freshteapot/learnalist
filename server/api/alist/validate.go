package alist

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/gookit/validate"
)

func Validate(aList Alist) error {
	var err error

	err = validateAListInfo(aList.Info)
	if err != nil {
		//err = errors.New(fmt.Sprintf("Failed to pass list info. %s", err.Error()))
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

	if info.From != nil {
		allowed := []string{"learnalist", "brainscape", "cram", "quizlet"}
		if !utils.StringArrayContains(allowed, info.From.Kind) {
			return i18n.ErrorInputSaveAlistFromKindNotSupported
		}

		v := validate.Struct(*info.From)

		//v.StringRule("Kind", "required|in:cram,brainscape,quizlet,learnalist")
		v.StringRule("RefUrl", "required")
		v.StringRule("ExtUuid", "required")

		if !v.Validate() {
			return ErrorListFromValid
		}

		if !WithFromCheckFromDomain(*info.From) {
			return ErrorListFromDomainMisMatch
		}

		if !WithFromCheckSharing(info) {
			return ErrorSharingNotAllowedWithFrom
		}
	}

	return err
}

func WithFromCanUpdate(want AlistInfo, current AlistInfo) bool {
	if want.From != current.From {
		return false
	}

	if current.From == nil {
		return true
	}

	return WithFromCheckSharing(want)
}

func WithFromCheckSharing(info AlistInfo) bool {
	// Defence, this shouldn't happen
	if info.From == nil {
		return true
	}

	if info.From.Kind == "learnalist" {
		return true
	}

	if info.SharedWith == aclKeys.NotShared {
		return true
	}

	return false
}

func WithFromCheckFromDomain(input openapi.AlistFrom) bool {
	allowed := map[string]string{
		"cram":       "cram.com",
		"brainscape": "brainscape.com",
		"quizlet":    "quizlet.com",
		"learnalist": "learnalist.net",
	}

	toTest := input.RefUrl
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	switch input.Kind {
	case "cram":
		fallthrough
	case "brainscape":
		fallthrough
	case "quizlet":
		fallthrough
	case "learnalist":
		return u.Host == allowed[input.Kind]
	case "localhost":
		return true
	default:
		return false
	}
}

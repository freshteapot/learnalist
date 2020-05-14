package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/gookit/validate"
)

// TypeV1 list type v1
type TypeV1 []string

func NewTypeV1() Alist {
	aList := Alist{
		Info: AlistInfo{
			Labels:   make([]string, 0),
			ListType: SimpleList,
			Interact: &Interact{
				Slideshow:   InteractDisabled,
				TotalRecall: InteractDisabled,
			},
		},
	}

	data := make(TypeV1, 0)
	aList.Data = data

	return aList
}

func validateTypeV1(aList Alist) error {
	hasError := false
	if !utils.IntArrayContains(ValidInteract, aList.Info.Interact.Slideshow) {
		hasError = true
	}

	if !utils.IntArrayContains(ValidInteract, aList.Info.Interact.TotalRecall) {
		hasError = true
	}

	// Little bit of a hack, due to the validate not working with slices.
	type itemv1 struct {
		Content string `json:"content" validate:"required"`
	}

	items := aList.Data.(TypeV1)
	for _, item := range items {
		a := itemv1{Content: item}
		v := validate.New(a)
		if !v.Validate() {
			hasError = true
		}
	}

	if hasError {
		return errors.New(i18n.ValidationAlistTypeV1)
	}
	return nil
}

func parseTypeV1(jsonBytes []byte) (TypeV1, error) {
	listData := new(TypeV1)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}

func parseInfoV1(info AlistInfo) (AlistInfo, error) {
	if info.Interact == nil {
		info.Interact = &Interact{
			Slideshow:   InteractDisabled,
			TotalRecall: InteractDisabled,
		}
	}

	return info, nil
}

package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
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
			SharedWith: aclKeys.NotShared,
		},
	}

	data := make(TypeV1, 0)
	aList.Data = data

	return aList
}

type mapToV1 struct{}

func NewMapToV1() AlistTypeMarshalJSON {
	return &mapToV1{}
}

func (m mapToV1) ParseInfo(info AlistInfo) (AlistInfo, error) {
	if info.Interact == nil {
		info.Interact = &Interact{
			Slideshow:   InteractDisabled,
			TotalRecall: InteractDisabled,
		}
	}
	return info, nil
}

func (m mapToV1) ParseData(jsonBytes []byte) (interface{}, error) {
	var listData TypeV1
	err := json.Unmarshal(jsonBytes, &listData)
	if err != nil {
		err = errors.New(i18n.ValidationErrorListV1)
	}
	return listData, err
}

func (m mapToV1) Enrich(aList Alist) Alist {
	return aList
}

func (m mapToV1) Validate(aList Alist) error {
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

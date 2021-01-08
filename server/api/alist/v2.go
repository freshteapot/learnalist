package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/gookit/validate"
)

// TypeV2Item Item in  TypeV2
type TypeV2Item struct {
	From string `json:"from" validate:"required"`
	To   string `json:"to" validate:"required"`
}

type TypeV2 []TypeV2Item

func NewTypeV2() Alist {
	aList := Alist{
		Info: AlistInfo{
			Labels:   make([]string, 0),
			ListType: FromToList,
			Interact: &Interact{
				TotalRecall: InteractDisabled,
			},
			SharedWith: aclKeys.NotShared,
		},
	}

	data := make(TypeV2, 0)
	aList.Data = data

	return aList
}

func (m mapToV2) Validate(aList Alist) error {
	hasError := false
	if !utils.IntArrayContains(ValidInteract, aList.Info.Interact.TotalRecall) {
		hasError = true
	}

	items := aList.Data.(TypeV2)
	for _, item := range items {
		v := validate.New(item)
		if !v.Validate() {
			hasError = true
		}
	}

	if hasError {
		return errors.New(i18n.ValidationAlistTypeV2)
	}
	return nil
}

type mapToV2 struct{}

func NewMapToV2() AlistTypeMarshalJSON {
	return &mapToV2{}
}

func (m mapToV2) ParseInfo(info AlistInfo) (AlistInfo, error) {
	if info.Interact == nil {
		info.Interact = &Interact{
			TotalRecall: InteractDisabled,
		}
	}

	return info, nil
}

func (m mapToV2) ParseData(jsonBytes []byte) (interface{}, error) {
	var listData TypeV2
	err := json.Unmarshal(jsonBytes, &listData)
	if err != nil {
		err = errors.New(i18n.ValidationErrorListV2)
	}
	return listData, err
}

func (m mapToV2) Enrich(aList Alist) Alist {
	return aList
}

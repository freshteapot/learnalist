package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/gookit/validate"
)

func NewTypeV4() Alist {
	aList := Alist{
		Info: AlistInfo{
			Labels:     make([]string, 0),
			ListType:   ContentAndUrl,
			SharedWith: aclKeys.NotShared,
		},
	}

	data := make(TypeV4, 0)
	aList.Data = data
	return aList
}

type TypeV4 []TypeV4Item

type TypeV4Item struct {
	Content string `json:"content" validate:"required"`
	Url     string `json:"url" validate:"required|url"`
}

type mapToV4 struct{}

func NewMapToV4() AlistTypeMarshalJSON {
	return &mapToV4{}
}

func (m mapToV4) Validate(aList Alist) error {
	hasError := false
	items := aList.Data.(TypeV4)
	for _, item := range items {
		v := validate.New(item)
		if !v.Validate() {
			hasError = true
		}
	}

	if hasError {
		return errors.New(i18n.ValidationAlistTypeV4)
	}
	return nil
}

func (m mapToV4) ParseInfo(info AlistInfo) (AlistInfo, error) {
	return info, nil
}

func (m mapToV4) ParseData(jsonBytes []byte) (interface{}, error) {
	var listData TypeV4
	err := json.Unmarshal(jsonBytes, &listData)
	if err != nil {
		err = errors.New(i18n.ValidationErrorListV4)
	}
	return listData, err
}

func (m mapToV4) Enrich(aList Alist) Alist {
	return aList
}

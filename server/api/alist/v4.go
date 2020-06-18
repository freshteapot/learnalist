package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/gookit/validate"
)

func NewTypeV4() Alist {
	aList := Alist{}

	aList.Info.ListType = ContentAndUrl
	data := make(TypeV4, 0)
	aList.Data = data
	return aList
}

type TypeV4 []TypeV4Item

type TypeV4Item struct {
	Content string `json:"content" validate:"required"`
	Url     string `json:"url" validate:"required|url"`
}

func parseTypeV4(jsonBytes []byte) (TypeV4, error) {
	listData := new(TypeV4)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}

func validateTypeV4(aList Alist) error {
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

type mapToV4 struct{}

func NewMapToV4() AlistTypeMarshalJSON {
	return &mapToV4{}
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

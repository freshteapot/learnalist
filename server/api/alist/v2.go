package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/gookit/validate"
)

// TypeV2Item Item in  TypeV2
type TypeV2Item struct {
	From string `json:"from" validate:"required"`
	To   string `json:"to" validate:"required"`
}

type TypeV2 []TypeV2Item

func NewTypeV2() *Alist {
	aList := &Alist{}

	aList.Info.ListType = FromToList
	data := make(TypeV2, 0)
	aList.Data = data

	labels := make([]string, 0)
	aList.Info.Labels = labels

	return aList
}

func parseTypeV2(jsonBytes []byte) (TypeV2, error) {
	listData := new(TypeV2)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}

func validateTypeV2(aList Alist) error {
	hasError := false
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

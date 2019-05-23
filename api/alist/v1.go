package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/gookit/validate"
)

// TypeV1 list type v1
type TypeV1 []string

func NewTypeV1() *Alist {
	aList := &Alist{}

	aList.Info.ListType = SimpleList
	data := make(TypeV1, 0)
	aList.Data = data

	labels := make([]string, 0)
	aList.Info.Labels = labels

	return aList
}

func validateTypeV1(aList Alist) error {
	// Little bit of a hack, due to the validate not working with slices.
	type itemv1 struct {
		Content string `json:"content" validate:"required"`
	}

	hasError := false
	items := aList.Data.(TypeV1)
	for _, item := range items {
		a := itemv1{Content: item}
		v := validate.New(a)
		if !v.Validate() { // validate ok
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

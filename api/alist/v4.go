package alist

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/gookit/validate"
)

func NewTypeV4() *Alist {
	aList := &Alist{}

	aList.Info.ListType = ContentAndUrl
	data := make(TypeV4, 0)
	aList.Data = data
	return aList
}

// TypeV3 Used for recording rowing machine times.
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
		if !v.Validate() { // validate ok
			hasError = true
		}
	}

	if hasError {
		return errors.New(i18n.ValidationAlistTypeV4)
	}
	return nil
}

package alist

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
)

func NewTypeV3() Alist {
	aList := Alist{}

	aList.Info.ListType = Concept2
	data := make(TypeV3, 0)
	aList.Data = data

	aList = enrichTypeV3(aList)
	return aList
}

// TypeV3 Used for recording rowing machine times.
type TypeV3 []TypeV3Item

type TypeV3Item struct {
	When    string    `json:"when"`
	Overall V3Split   `json:"overall"`
	Splits  []V3Split `json:"splits"`
}

type V3Split struct {
	Time     string `json:"time"`
	Distance int    `json:"distance"`
	Spm      int    `json:"spm"`
	P500     string `json:"p500"`
}

/*
"when": "2019-05-06",
    "overall": {
      "time": "7:15.9",
      "distance": "2000m",
      "spm": "28",
      "p500": "1:48.9",
    },
    "splits": [
      {
        "time": "1.46.4",
        "distance": "500",
        "spm": "29",
        "p500": "500",
      },
      {
*/

func enrichTypeV3(aList Alist) Alist {
	labels := aList.Info.Labels
	if !utils.StringArrayContains(labels, "rowing") {
		labels = append(labels, "rowing")
	}
	if !utils.StringArrayContains(labels, "concept2") {
		labels = append(labels, "concept2")
	}
	aList.Info.Labels = labels
	return aList
}

func ParseTypeV3(jsonBytes []byte) (TypeV3, error) {
	listData := new(TypeV3)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}

func ValidateTypeV3(typeV3 TypeV3) error {
	var err error
	var feedback = errors.New(i18n.ValidationAlistTypeV3)

	for _, item := range typeV3 {

		err = ValidateTypeV3When(item.When)
		if err != nil {
			return feedback
		}

		// overall
		err = ValidateTypeV3Split(item.Overall)
		if err != nil {
			return feedback
		}

		// Splits
		for _, split := range item.Splits {
			err = ValidateTypeV3Split(split)
			if err != nil {
				return feedback
			}
		}
	}

	return nil
}

func ValidateTypeV3Split(split V3Split) error {
	var err error
	err = ValidateTypeV3Time(split.Time)
	if err != nil {
		return err
	}

	err = ValidateTypeV3Distance(split.Distance)
	if err != nil {
		return err
	}

	err = ValidateTypeV3Spm(split.Spm)
	if err != nil {
		return err
	}

	err = ValidateTypeV3P500(split.P500)
	if err != nil {
		return err
	}
	return nil
}

func ValidateTypeV3Distance(input int) error {
	var err error
	if input == 0 {
		return errors.New("Should not be empty.")
	}
	return err
}

func ValidateTypeV3Spm(input int) error {
	var err error
	if input < 10 || input > 50 {
		return errors.New("Stroke per minute should be between the range 10 and 50.")
	}
	return err
}

func ValidateTypeV3Time(input string) error {
	var err error
	if input == "" {
		return errors.New("Should not be empty.")
	}

	if !strings.Contains(input, ":") {
		return errors.New("Is not valid format.")
	}

	if !strings.Contains(input, ".") {
		return errors.New("Is not valid format.")
	}
	// TODO maybe do a better check
	return err
}

func ValidateTypeV3When(input string) error {
	var err error
	if input == "" {
		return errors.New("When should be YYYY-MM-DD.")
	}
	parsed, err := dateparse.ParseAny(input)

	if parsed.Format("2006-01-02") != input {
		return errors.New("When should be YYYY-MM-DD.")
	}
	return err
}

func ValidateTypeV3P500(input string) error {
	return ValidateTypeV3Time(input)
}

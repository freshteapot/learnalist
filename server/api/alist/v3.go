package alist

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
)

func NewTypeV3() *Alist {
	aList := &Alist{}

	aList.Info.ListType = Concept2
	data := make(TypeV3, 0)
	aList.Data = data

	labels := []string{
		"rowing",
		"concept2",
	}

	aList.Info.Labels = labels

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

func parseTypeV3(jsonBytes []byte) (TypeV3, error) {
	listData := new(TypeV3)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}

func validateTypeV3(aList Alist) error {
	var err error
	var feedback = errors.New(i18n.ValidationAlistTypeV3)

	typeV3 := aList.Data.(TypeV3)
	for _, item := range typeV3 {

		err = validateTypeV3When(item.When)
		if err != nil {
			return feedback
		}

		err = validateTypeV3Split(item.Overall)
		if err != nil {
			return feedback
		}

		for _, split := range item.Splits {
			err = validateTypeV3Split(split)
			if err != nil {
				return feedback
			}
		}
	}

	return nil
}

func validateTypeV3Split(split V3Split) error {
	var err error
	err = validateTypeV3Time(split.Time)
	if err != nil {
		return err
	}

	err = validateTypeV3Distance(split.Distance)
	if err != nil {
		return err
	}

	err = validateTypeV3Spm(split.Spm)
	if err != nil {
		return err
	}

	err = validateTypeV3P500(split.P500)
	if err != nil {
		return err
	}
	return nil
}

func validateTypeV3Distance(input int) error {
	var err error
	if input == 0 {
		return errors.New("Distance should not be empty.")
	}
	return err
}

func validateTypeV3Spm(input int) error {
	var err error
	if input < 10 || input > 50 {
		return errors.New("Stroke per minute should be between the range 10 and 50.")
	}
	return err
}

func validateTypeV3Time(input string) error {
	var err error
	if input == "" {
		return errors.New("Time should not be empty.")
	}

	if !strings.Contains(input, ":") {
		return errors.New("Time is not valid format.")
	}

	if !strings.Contains(input, ".") {
		return errors.New("Time is not valid format.")
	}
	// TODO maybe do a better check
	return err
}

func validateTypeV3When(input string) error {
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

func validateTypeV3P500(input string) error {
	var err error
	if input == "" {
		return errors.New("Per 500 should not be empty.")
	}

	if !strings.Contains(input, ":") {
		return errors.New("Per 500 is not valid format.")
	}

	if !strings.Contains(input, ".") {
		return errors.New("Per 500 is not valid format.")
	}
	// TODO maybe do a better check
	return err
}

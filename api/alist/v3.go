package alist

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/araddon/dateparse"
	"github.com/freshteapot/learnalist-api/api/utils"
)

// TypeV3 Used for recording rowing machine times.
type TypeV3 struct {
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
	var feedbackMessage string
	var feedback []string

	typeV3 := aList.Data.(TypeV3)
	err = validateTypeV3When(typeV3.When)
	if err != nil {
		feedback = append(feedback, err.Error())
	}

	err = validateTypeV3Time(typeV3.Overall.Time)
	if err != nil {
		feedback = append(feedback, err.Error())
	}

	err = validateTypeV3Distance(typeV3.Overall.Distance)
	if err != nil {
		feedback = append(feedback, err.Error())
	}

	err = validateTypeV3Spm(typeV3.Overall.Spm)
	if err != nil {
		feedback = append(feedback, err.Error())
	}

	err = validateTypeV3P500(typeV3.Overall.P500)
	if err != nil {
		feedback = append(feedback, err.Error())
	}

	// TODO should we validate the splits
	if len(feedback) != 0 {
		feedbackMessage = strings.Join(feedback, "\n")
		err = errors.New(feedbackMessage)
	}

	return err
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

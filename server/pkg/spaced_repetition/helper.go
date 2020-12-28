package spaced_repetition

import (
	"encoding/json"
	"time"
)

func CheckNext(entry SpacedRepetitionEntry, err error) (interface{}, error) {
	var body interface{}
	if err != nil {
		return body, err
	}

	if !time.Now().UTC().After(entry.WhenNext) {
		return body, ErrFoundNotTime
	}

	json.Unmarshal([]byte(entry.Body), &body)
	return body, nil
}

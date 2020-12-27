package spaced_repetition

import (
	"encoding/json"
	"time"
)

// TODO does this need to be part of the struct / interface?
func CheckNext(entry SpacedRepetitionEntry, err error) (interface{}, error) {
	var body interface{}
	if err != nil {
		if err == ErrNotFound {
			return body, ErrNotFound
		}
		return body, err
	}

	if !time.Now().UTC().After(entry.WhenNext) {
		return body, ErrFoundNotTime
	}

	json.Unmarshal([]byte(entry.Body), &body)
	return body, nil
}

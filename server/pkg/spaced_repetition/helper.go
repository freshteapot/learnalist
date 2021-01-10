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

func DefaultSettingsV1() HTTPRequestInputSettings {
	now := time.Now().UTC()
	whenNext := now.Add(Threshold0)
	return HTTPRequestInputSettings{
		Level:    Level0,
		Created:  now.Format(time.RFC3339),
		WhenNext: whenNext.Format(time.RFC3339),
		ExtID:    "",
	}
}

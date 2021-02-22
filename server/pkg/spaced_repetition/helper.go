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

func DefaultSettingsV1(now time.Time) HTTPRequestInputSettings {
	whenNext := now.Add(Threshold0)
	return HTTPRequestInputSettings{
		Level:    Level0,
		Created:  now.Format(time.RFC3339),
		WhenNext: whenNext.Format(time.RFC3339),
		ExtID:    "",
	}
}

func DefaultSettingsV2(now time.Time) HTTPRequestInputSettingsV2 {
	baseSettings := DefaultSettingsV1(now)

	settings := HTTPRequestInputSettingsV2{}
	settings.Created = baseSettings.Created
	settings.WhenNext = baseSettings.WhenNext
	settings.Level = baseSettings.Level
	settings.ExtID = baseSettings.ExtID
	settings.Show = "from"
	return settings
}

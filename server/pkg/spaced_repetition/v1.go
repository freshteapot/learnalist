package spaced_repetition

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"time"
)

type ItemInputV1 struct {
	entry *HttpRequestInputV1
}

func V1FromPOST(input []byte) ItemInputV1 {
	item := ItemInputV1{}

	json.Unmarshal(input, &item.entry)

	b, _ := json.Marshal(item.entry.Data)
	hash := fmt.Sprintf("%x", sha1.Sum(b))
	item.entry.UUID = hash

	item.entry.Settings.Level = Level_0
	whenNext := time.Now().Add(time.Hour * 1).UTC()
	item.entry.Settings.WhenNext = whenNext.Format(time.RFC3339)
	return item
}

func V1FromDB(input string) ItemInputV1 {
	item := ItemInputV1{}

	json.Unmarshal([]byte(input), &item.entry)
	return item
}

func (item ItemInputV1) String() string {
	b, _ := json.Marshal(item.entry)
	return string(b)
}

func (item ItemInputV1) UUID() string {
	return item.entry.UUID
}

func (item ItemInputV1) WhenNext() time.Time {
	t, _ := time.Parse(time.RFC3339, item.entry.Settings.WhenNext)
	return t
}

func (item ItemInputV1) DecrThreshold() {
	for _, find := range decrThresholds {
		if find.Match == item.entry.Settings.Level {
			whenNext := time.Now().UTC().Add(find.Threshold)
			item.entry.Settings.WhenNext = whenNext.Format(time.RFC3339)
			item.entry.Settings.Level = find.Level
			break
		}
	}
}

func (item ItemInputV1) IncrThreshold() {
	for _, find := range incrThresholds {
		if find.Match == item.entry.Settings.Level {
			whenNext := time.Now().UTC().Add(find.Threshold)
			item.entry.Settings.WhenNext = whenNext.Format(time.RFC3339)
			item.entry.Settings.Level = find.Level
			break
		}
	}
}

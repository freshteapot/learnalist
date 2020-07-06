package spaced_repetition

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"time"
)

type ItemInputV2 struct {
	entry *HttpRequestInputV2
}

func V2FromPOST(input []byte) ItemInputV2 {
	item := ItemInputV2{}

	json.Unmarshal(input, &item.entry)

	b, _ := json.Marshal(item.entry.Data)
	hash := fmt.Sprintf("%x", sha1.Sum(b))
	item.entry.UUID = hash

	item.entry.Settings.Level = Level_0
	whenNext := time.Now().Add(time.Hour * 1).UTC()
	item.entry.Settings.WhenNext = whenNext.Format(time.RFC3339)
	return item
}

func V2FromDB(input string) ItemInputV2 {
	item := ItemInputV2{}

	json.Unmarshal([]byte(input), &item.entry)
	return item
}

func (item ItemInputV2) String() string {
	fmt.Println(item.entry.Settings.WhenNext)
	b, _ := json.Marshal(item.entry)
	return string(b)
}

func (item ItemInputV2) UUID() string {
	return item.entry.UUID
}

func (item ItemInputV2) WhenNext() time.Time {
	t, _ := time.Parse(time.RFC3339, item.entry.Settings.WhenNext)
	return t
}

func (item ItemInputV2) DecrThreshold() {

	for _, find := range decrThresholds {
		if find.Match == item.entry.Settings.Level {
			whenNext := time.Now().UTC().Add(find.Threshold)
			item.entry.Settings.WhenNext = whenNext.Format(time.RFC3339)
			item.entry.Settings.Level = find.Level
			break
		}
	}
}

func (item ItemInputV2) IncrThreshold() {
	for _, find := range incrThresholds {
		if find.Match == item.entry.Settings.Level {
			whenNext := time.Now().UTC().Add(find.Threshold)
			item.entry.Settings.WhenNext = whenNext.Format(time.RFC3339)
			item.entry.Settings.Level = find.Level
			break
		}
	}
}

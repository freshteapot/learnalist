package spaced_repetition

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

type ItemInputV1 struct {
	entry *HTTPRequestInputV1
}

func V1FromPOST(input []byte, settings HTTPRequestInputSettings) ItemInputV1 {
	item := ItemInputV1{}

	json.Unmarshal(input, &item.entry)

	b, _ := json.Marshal(item.entry.Data)
	hash := fmt.Sprintf("%x", sha1.Sum(b))
	item.entry.UUID = hash
	item.entry.HTTPRequestInput.Kind = alist.SimpleList
	item.entry.HTTPRequestInput.Show = item.entry.Data
	item.entry.Settings = settings
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

func (item ItemInputV1) Created() time.Time {
	t, _ := time.Parse(time.RFC3339, item.entry.Settings.Created)
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

func (item ItemInputV1) SetExtID(extID string) {
	item.entry.Settings.ExtID = extID
}

func (item ItemInputV1) Reset(now time.Time) {
	item.entry.Settings.Level = Level0
	//now := time.Now().UTC()
	whenNext := now.Add(Threshold0)
	item.entry.Settings.Created = now.Format(time.RFC3339)
	item.entry.Settings.WhenNext = whenNext.Format(time.RFC3339)
}

package spaced_repetition

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

type ItemInputV2 struct {
	entry *HTTPRequestInputV2
}

// TODO do we add error to this to validate?
// TODO override show, confirm what we set it to is valid or reject
// TODO valide settings.show
func V2FromPOST(input []byte, settings HTTPRequestInputSettings) ItemInputV2 {
	item := ItemInputV2{}
	// TODO set show
	json.Unmarshal(input, &item.entry)

	b, _ := json.Marshal(item.entry.Data)
	hash := fmt.Sprintf("%x", sha1.Sum(b))
	item.entry.UUID = hash

	item.entry.HTTPRequestInput.Kind = alist.FromToList
	//item.entry.HTTPRequestInput.Show = item.entry.Data

	// TODO confirm show comes thru via the POST
	item.entry.Settings.Level = settings.Level
	item.entry.Settings.Created = settings.Created
	item.entry.Settings.WhenNext = settings.WhenNext
	item.entry.Settings.ExtID = settings.ExtID
	return item
}

func V2FromDB(input string) ItemInputV2 {
	item := ItemInputV2{}

	json.Unmarshal([]byte(input), &item.entry)
	return item
}

func (item ItemInputV2) String() string {
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

func (item ItemInputV2) Created() time.Time {
	t, _ := time.Parse(time.RFC3339, item.entry.Settings.Created)
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

func (item ItemInputV2) SetExtID(extID string) {
	item.entry.Settings.ExtID = extID
}

func (item ItemInputV2) Reset(now time.Time) {
	item.entry.Settings.Level = Level0
	//now := time.Now().UTC()
	whenNext := now.Add(Threshold0)
	item.entry.Settings.Created = now.Format(time.RFC3339)
	item.entry.Settings.WhenNext = whenNext.Format(time.RFC3339)
}

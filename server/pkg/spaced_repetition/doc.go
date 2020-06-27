package spaced_repetition

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type HttpRequestInputKind struct {
	Kind string `json:"kind"`
}

type HttpRequestViewed struct {
	UUID   string `json:"uuid"`
	Action string `json:"action"`
}

type HttpRequestInput struct {
	Show string `json:"show"`
	Kind string `json:"kind"`
	UUID string `json:"uuid"`
}

type HttpRequestInputV1 struct {
	HttpRequestInput
	Data     string                   `json:"data"`
	Settings HttpRequestInputSettings `json:"settings"`
}

type HttpRequestInputV2 struct {
	HttpRequestInput
	Data     HttpRequestInputV2Item     `json:"data"`
	Settings HttpRequestInputSettingsV2 `json:"settings"`
}

type HttpRequestInputV2Item struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// TODO add level here
type HttpRequestInputSettings struct {
	Level    string `json:"level"`
	WhenNext string `json:"when_next"`
}
type HttpRequestInputSettingsV2 struct {
	HttpRequestInputSettings
	Show string `json:"show"`
}

type SpacedRepetitionItem struct {
	UUID     string    `db:"uuid"`
	Body     string    `db:"body"`
	UserUUID string    `db:"user_uuid"`
	WhenNext time.Time `db:"when_next"`
}

const (
	Level_0 = "0"
	Level_1 = "1"
	Level_2 = "2"
	Level_3 = "3"
	Level_4 = "4"
	Level_5 = "5"
	Level_6 = "6"
	Level_7 = "7"
	Level_8 = "8"
	Level_9 = "9"

	TIME_DAY    = 24 * time.Hour
	THRESHOLD_0 = time.Duration(time.Hour * 1)
	THRESHOLD_1 = time.Duration(time.Hour * 3)
	THRESHOLD_2 = time.Duration(time.Hour * 8)
	THRESHOLD_3 = time.Duration(TIME_DAY * 1)
	THRESHOLD_4 = time.Duration(TIME_DAY * 3)
	THRESHOLD_5 = time.Duration(TIME_DAY * 7)
	THRESHOLD_6 = time.Duration(TIME_DAY * 14)
	THRESHOLD_7 = time.Duration(TIME_DAY * 30)
	THRESHOLD_8 = time.Duration(TIME_DAY * 60)
	THRESHOLD_9 = time.Duration(TIME_DAY * 120)

	SQL_SAVE_ITEM   = `INSERT INTO spaced_repetition(uuid, body, user_uuid, when_next) values(?, ?, ?, ?) ON CONFLICT (spaced_repetition.user_uuid, spaced_repetition.uuid) DO UPDATE SET body=?, when_next=?`
	SQL_DELETE_ITEM = `DELETE FROM spaced_repetition WHERE uuid=? AND user_uuid=?`
	SQL_GET_ITEM    = `SELECT * FROM spaced_repetition WHERE uuid=? AND user_uuid=?`
	// TODO add order by when_next
	// ADD index when_next
	SQL_GET_ALL  = `SELECT body FROM spaced_repetition WHERE user_uuid=? ORDER BY when_next`
	SQL_GET_NEXT = `SELECT * FROM spaced_repetition WHERE user_uuid=? ORDER BY when_next LIMIT 1`
)

var incrThresholds = []struct {
	Match     string
	Level     string
	Threshold time.Duration
}{
	{
		Match:     Level_0,
		Level:     Level_1,
		Threshold: THRESHOLD_1,
	},
	{
		Match:     Level_1,
		Level:     Level_2,
		Threshold: THRESHOLD_2,
	},
	{
		Match:     Level_2,
		Level:     Level_3,
		Threshold: THRESHOLD_3,
	},
	{
		Match:     Level_3,
		Level:     Level_4,
		Threshold: THRESHOLD_4,
	},
	{
		Match:     Level_4,
		Level:     Level_5,
		Threshold: THRESHOLD_5,
	},
	{
		Match:     Level_5,
		Level:     Level_6,
		Threshold: THRESHOLD_6,
	},
	{
		Match:     Level_6,
		Level:     Level_7,
		Threshold: THRESHOLD_7,
	},
	{
		Match:     Level_7,
		Level:     Level_8,
		Threshold: THRESHOLD_8,
	},
	{
		Match:     Level_8,
		Level:     Level_9,
		Threshold: THRESHOLD_9,
	},
}

var decrThresholds = []struct {
	Match     string
	Level     string
	Threshold time.Duration
}{
	{
		Match:     Level_0,
		Level:     Level_0,
		Threshold: THRESHOLD_0,
	},
	{
		Match:     Level_1,
		Level:     Level_0,
		Threshold: THRESHOLD_0,
	},
	{
		Match:     Level_2,
		Level:     Level_1,
		Threshold: THRESHOLD_1,
	},
	{
		Match:     Level_3,
		Level:     Level_2,
		Threshold: THRESHOLD_2,
	},
	{
		Match:     Level_4,
		Level:     Level_3,
		Threshold: THRESHOLD_3,
	},
	{
		Match:     Level_5,
		Level:     Level_4,
		Threshold: THRESHOLD_4,
	},
	{
		Match:     Level_6,
		Level:     Level_5,
		Threshold: THRESHOLD_5,
	},
	{
		Match:     Level_7,
		Level:     Level_6,
		Threshold: THRESHOLD_6,
	},
	{
		Match:     Level_8,
		Level:     Level_7,
		Threshold: THRESHOLD_7,
	},
	{
		Match:     Level_9,
		Level:     Level_8,
		Threshold: THRESHOLD_8,
	},
}

type service struct {
	db *sqlx.DB
}

type ItemInput interface {
	UUID() string
	String() string
	WhenNext() time.Time
	IncrThreshold()
	DecrThreshold()
}

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

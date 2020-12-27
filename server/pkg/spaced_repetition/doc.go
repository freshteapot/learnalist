package spaced_repetition

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

type HTTPRequestInputKind struct {
	Kind string `json:"kind"`
}

type HTTPRequestViewed struct {
	UUID   string `json:"uuid"`
	Action string `json:"action"`
}

type HTTPRequestInput struct {
	Show string `json:"show"`
	Kind string `json:"kind"`
	UUID string `json:"uuid"`
}

type HTTPRequestInputV1 struct {
	HTTPRequestInput
	Data     string                   `json:"data"`
	Settings HTTPRequestInputSettings `json:"settings"`
}

type HTTPRequestInputV2 struct {
	HTTPRequestInput
	Data     HTTPRequestInputV2Item     `json:"data"`
	Settings HTTPRequestInputSettingsV2 `json:"settings"`
}

type HTTPRequestInputV2Item struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type HTTPRequestInputSettings struct {
	Level    string `json:"level"`
	WhenNext string `json:"when_next"`
	Created  string `json:"created"`
}
type HTTPRequestInputSettingsV2 struct {
	HTTPRequestInputSettings
	Show string `json:"show"`
}

type SpacedRepetitionEntry struct {
	UUID     string    `json:"uuid" db:"uuid"`
	Body     string    `json:"body" db:"body"`
	UserUUID string    `json:"user_uuid" db:"user_uuid"`
	WhenNext time.Time `json:"when_next" db:"when_next"`
	Created  time.Time `json:"created" db:"created"`
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

	ActionIncrement = "incr"
	ActionDecrement = "decr"

	SQL_SAVE_ITEM              = `INSERT INTO spaced_repetition(uuid, body, user_uuid, when_next, created) values(?, ?, ?, ?, ?)`
	SQL_SAVE_ITEM_AUTO_UPDATED = `INSERT INTO spaced_repetition(uuid, body, user_uuid, when_next) values(?, ?, ?, ?) ON CONFLICT (spaced_repetition.user_uuid, spaced_repetition.uuid) DO UPDATE SET body=?, when_next=?`
	SQL_DELETE_ITEM            = `DELETE FROM spaced_repetition WHERE uuid=? AND user_uuid=?`
	SQL_GET_ITEM               = `SELECT * FROM spaced_repetition WHERE uuid=? AND user_uuid=?`
	SQL_GET_ALL                = `SELECT body FROM spaced_repetition WHERE user_uuid=? ORDER BY when_next`
	SQL_GET_NEXT               = `SELECT * FROM spaced_repetition WHERE user_uuid=? ORDER BY when_next LIMIT 1`
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

type SpacedRepetitionService struct {
	repo       SpacedRepetitionRepository
	logContext logrus.FieldLogger
}

type SpacedRepetitionRepository interface {
	GetNext(userUUID string) (SpacedRepetitionEntry, error)
	GetEntry(userUUID string, UUID string) (interface{}, error)
	GetEntries(userUUID string) ([]interface{}, error)
	SaveEntry(entry SpacedRepetitionEntry) error
	UpdateEntry(entry SpacedRepetitionEntry) error
	DeleteEntry(userUUID string, UUID string) error
}

type ItemInput interface {
	UUID() string
	String() string
	WhenNext() time.Time
	Created() time.Time
	IncrThreshold()
	DecrThreshold()
}

var (
	ErrNotFound                    = errors.New("not.found")
	ErrFoundNotTime                = errors.New("found.not.time")
	ErrSpacedRepetitionEntryExists = errors.New("item.exists")
)

// Event specific
var (
	EventKindNew     = "new"
	EventKindViewed  = "viewed"
	EventKindDeleted = "deleted"
)

type EventSpacedRepetition struct {
	Kind   string                `json:"kind"`
	Action string                `json:"action,omitempty"`
	Data   SpacedRepetitionEntry `json:"data"`
}

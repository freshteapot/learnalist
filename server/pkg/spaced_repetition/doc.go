package spaced_repetition

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
)

type HTTPRequestInputKind struct {
	Kind string `json:"kind"`
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

// Should we move this to openapi?
type HTTPRequestInputSettings struct {
	Level    string `json:"level"`
	WhenNext string `json:"when_next"`
	Created  string `json:"created"`
	ExtID    string `json:"ext_id,omitempty"` // ext_id used by dripfeed, at first
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
	Level0 = "0"
	Level1 = "1"
	Level2 = "2"
	Level3 = "3"
	Level4 = "4"
	Level5 = "5"
	Level6 = "6"
	Level7 = "7"
	Level8 = "8"
	Level9 = "9"

	TimeDay    = 24 * time.Hour
	Threshold0 = time.Duration(time.Hour * 1)
	Threshold1 = time.Duration(time.Hour * 3)
	Threshold2 = time.Duration(time.Hour * 8)
	Threshold3 = time.Duration(TimeDay * 1)
	Threshold4 = time.Duration(TimeDay * 3)
	Threshold5 = time.Duration(TimeDay * 7)
	Threshold6 = time.Duration(TimeDay * 14)
	Threshold7 = time.Duration(TimeDay * 30)
	Threshold8 = time.Duration(TimeDay * 60)
	Threshold9 = time.Duration(TimeDay * 120)

	ActionIncrement = "incr"
	ActionDecrement = "decr"
)

var incrThresholds = []struct {
	Match     string
	Level     string
	Threshold time.Duration
}{
	{
		Match:     Level0,
		Level:     Level1,
		Threshold: Threshold1,
	},
	{
		Match:     Level1,
		Level:     Level2,
		Threshold: Threshold2,
	},
	{
		Match:     Level2,
		Level:     Level3,
		Threshold: Threshold3,
	},
	{
		Match:     Level3,
		Level:     Level4,
		Threshold: Threshold4,
	},
	{
		Match:     Level4,
		Level:     Level5,
		Threshold: Threshold5,
	},
	{
		Match:     Level5,
		Level:     Level6,
		Threshold: Threshold6,
	},
	{
		Match:     Level6,
		Level:     Level7,
		Threshold: Threshold7,
	},
	{
		Match:     Level7,
		Level:     Level8,
		Threshold: Threshold8,
	},
	{
		Match:     Level8,
		Level:     Level9,
		Threshold: Threshold9,
	},
}

var decrThresholds = []struct {
	Match     string
	Level     string
	Threshold time.Duration
}{
	{
		Match:     Level0,
		Level:     Level0,
		Threshold: Threshold0,
	},
	{
		Match:     Level1,
		Level:     Level0,
		Threshold: Threshold0,
	},
	{
		Match:     Level2,
		Level:     Level1,
		Threshold: Threshold1,
	},
	{
		Match:     Level3,
		Level:     Level2,
		Threshold: Threshold2,
	},
	{
		Match:     Level4,
		Level:     Level3,
		Threshold: Threshold3,
	},
	{
		Match:     Level5,
		Level:     Level4,
		Threshold: Threshold4,
	},
	{
		Match:     Level6,
		Level:     Level5,
		Threshold: Threshold5,
	},
	{
		Match:     Level7,
		Level:     Level6,
		Threshold: Threshold6,
	},
	{
		Match:     Level8,
		Level:     Level7,
		Threshold: Threshold7,
	},
	{
		Match:     Level9,
		Level:     Level8,
		Threshold: Threshold8,
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
	DeleteByUser(userUUID string) error
}

type ItemInput interface {
	UUID() string
	String() string
	WhenNext() time.Time
	Created() time.Time
	IncrThreshold()
	DecrThreshold()
	SetExtID(extID string) // TODO add this
	Reset(now time.Time)   // TODO really? maybe pass in time
}

var (
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

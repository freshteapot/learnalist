package challenge

import (
	"errors"
	"time"
)

// Users are required for the name
// Could start without name for now.
type HttpChallengePlankRecords struct {
	UUID    string                 `json:"uuid"`
	Users   []ChallengePlankUsers  `json:"users"`
	Records []ChallengePlankRecord `json:"records"`
}

// TODO user doesnt have a name.
// TODO maybe add it to JOIN
type ChallengePlankUsers struct {
	UserUUID string `json:"user_uuid"`
	Name     string `json:"name"`
}

type ChallengePlankRecord struct {
	UserUUID         string `json:"user_uuid"`
	UUID             string `json:"uuid"`
	ShowIntervals    bool   `json:"showIntervals"`
	IntervalTime     int    `json:"intervalTime"`
	BeginningTime    int64  `json:"beginningTime"`
	CurrentTime      int64  `json:"currentTime"`
	TimerNow         int    `json:"timerNow"`
	IntervalTimerNow int    `json:"intervalTimerNow"`
	Laps             int    `json:"laps"`
}

type ChallengeInfo struct {
	UUID        string                 `json:"uuid"`
	Kind        string                 `json:"kind"`
	Description string                 `json:"description"`
	Created     string                 `json:"created"`
	Users       []ChallengePlankUsers  `json:"users,omitempty"`
	Records     []ChallengePlankRecord `json:"records,omitempty"`
}

type ChallengeInfoDB struct {
	UUID     string    `db:"uuid"`
	UserUUID string    `db:"user_uuid"`
	Body     string    `db:"body"`
	Created  time.Time `db:"created"`
}

type ChallengeShortInfoDB struct {
	UUID        string    `db:"uuid"`
	Description string    `db:"description"`
	Kind        string    `db:"kind"`
	Created     time.Time `db:"created"`
}

type ChallengeShortInfo struct {
	UUID        string `json:"uuid"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
	Created     string `json:"created"`
}

type ChallengeRepository interface {
	GetChallengesByUser(userUUID string) ([]ChallengeShortInfo, error)
	Join(UUID string, userUUID string) error
	Leave(UUID string, userUUID string) error
	Create(userUUID string, challenge ChallengeInfo) error
	Get(UUID string) (ChallengeInfo, error)
	Delete(UUID string) error
	DeleteUser(userUUID string) error
	AddRecord(UUID string, extUUID string, userUUID string) error
	DeleteRecord(extUUID string, userUUID string) error
}

type EventChallengeDoneEntry struct {
	Kind     string      `json:"kind"`
	UUID     string      `json:"uuid"`
	UserUUID string      `json:"user_uuid"`
	Data     interface{} `json:"data"`
}

// Event specific
var (
	EventChallengeDone        = "challenge.done"
	EventKindPlank            = "plank"
	EventKindSpacedRepetition = "srs"
)

type EventEntry struct {
	Kind string                  `json:"kind"`
	Data EventChallengeDoneEntry `json:"data"`
}

var (
	ErrNotFound = errors.New("not.found")
)

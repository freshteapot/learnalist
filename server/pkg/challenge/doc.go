package challenge

import (
	"errors"
	"time"
)

// Copy of plank.HttpRequestInput
type ChallengePlankRecord struct {
	UUID             string `json:"uuid"`
	ShowIntervals    bool   `json:"showIntervals"`
	IntervalTime     int    `json:"intervalTime"`
	BeginningTime    int64  `json:"beginningTime"`
	CurrentTime      int64  `json:"currentTime"`
	TimerNow         int    `json:"timerNow"`
	IntervalTimerNow int    `json:"intervalTimerNow"`
	Laps             int    `json:"laps"`
}

type ChallengeBody struct {
	UUID        string `json:"uuid"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Created     string `json:"created"`
}

type ChallengeEntry struct {
	UUID     string    `db:"uuid"`
	UserUUID string    `db:"user_uuid"`
	Body     string    `db:"body"`
	Created  time.Time `db:"created"`
}

type ChallengeRepository interface {
	GetChallengesByUser(userUUID string) ([]ChallengeBody, error)
	Join(UUID string, userUUID string) error
	Leave(UUID string, userUUID string) error
	Create(challenge ChallengeEntry) error
	Get(UUID string) (ChallengeBody, error)
	Delete(UUID string) error
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

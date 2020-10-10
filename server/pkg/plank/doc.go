package plank

import (
	"errors"
	"time"
)

var (
	ErrNotFound    = errors.New("not.found")
	ErrEntryExists = errors.New("item.exists")
)

type Repository interface {
	SaveEntry(entry Entry) error
	// Return in time order, latest first
	History(userUUID string) ([]HttpRequestInput, error)
	DeleteEntry(userUUID string, UUID string) error
	DeleteEntriesByUser(userUUID string) error
}

type HttpRequestInput struct {
	UUID             string `json:"uuid,omitempty"`
	ShowIntervals    bool   `json:"showIntervals"`
	IntervalTime     int    `json:"intervalTime"`
	BeginningTime    int64  `json:"beginningTime"`
	CurrentTime      int64  `json:"currentTime"`
	TimerNow         int    `json:"timerNow"`
	IntervalTimerNow int    `json:"intervalTimerNow"`
	Laps             int    `json:"laps"`
}

// Might need to evole this when I eventually move from sqlite
type Entry struct {
	UUID     string           `json:"uuid" db:"uuid"`
	UserUUID string           `json:"user_uuid" db:"user_uuid"`
	Body     HttpRequestInput `json:"body" db:"body"`
	Created  time.Time        `json:"created" db:"created"`
}

var (
	EventApiPlank    = "api.plank"
	EventKindNew     = "new"
	EventKindDeleted = "deleted"
)

type EventPlank struct {
	Kind     string           `json:"kind"`
	Data     HttpRequestInput `json:"data"`
	UserUUID string           `json:"user_uuid"`
}

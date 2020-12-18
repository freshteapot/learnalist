package plank

import (
	"errors"
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
)

var (
	ErrNotFound    = errors.New("not.found")
	ErrEntryExists = errors.New("item.exists")
)

type PlankRepository interface {
	GetEntry(UUID string, userUUID string) (openapi.Plank, error)
	SaveEntry(entry Entry) error
	// Return in time order, latest first
	History(userUUID string) ([]openapi.Plank, error)
	DeleteEntry(UUID string, userUUID string) error
	DeleteEntriesByUser(userUUID string) error
}

// Might need to evole this when I eventually move from sqlite
type Entry struct {
	UUID     string        `json:"uuid" db:"uuid"`
	UserUUID string        `json:"user_uuid" db:"user_uuid"`
	Body     openapi.Plank `json:"body" db:"body"`
	Created  time.Time     `json:"created" db:"created"`
}

var (
	EventKindDeleted = "deleted"
)

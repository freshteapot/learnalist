package spaced_repetition

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type HttpRequestInput struct {
	Show     string                   `json:"show"`
	Data     HttpRequestInputData     `json:"data"`
	Settings HttpRequestInputSettings `json:"settings"`
	Kind     string                   `json:"kind"`
}

type HttpRequestInputData struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// TODO add level here
type HttpRequestInputSettings struct {
	Show     string `json:"show"`
	Level    string `json:"level"`
	WhenNext string `json:"when_next"`
}

type SpacedRepetitionItem struct {
	UUID     string    `db:"uuid"`
	Body     string    `db:"body"`
	UserUUID string    `db:"user_uuid"`
	WhenNext time.Time `db:"when_next"`
}

const (
	Level_0         = "0"
	Level_1         = "1"
	Level_2         = "2"
	Level_3         = "3"
	Level_4         = "4"
	Level_5         = "4"
	SQL_SAVE_ITEM   = `INSERT INTO spaced_repetition(uuid, body, user_uuid, when_next) values(?, ?, ?, ?)`
	SQL_DELETE_ITEM = `DELETE FROM spaced_repetition WHERE uuid=? AND user_uuid=?`
	// TODO add order by when_next
	// ADD index when_next
	SQL_GET_ALL  = `SELECT body FROM spaced_repetition WHERE user_uuid=? ORDER BY when_next`
	SQL_GET_NEXT = `SELECT * FROM spaced_repetition WHERE user_uuid=? ORDER BY when_next LIMIT 1`
)

type service struct {
	db *sqlx.DB
}

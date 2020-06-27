package spaced_repetition

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type HttpRequestInputKind struct {
	Kind string `json:"kind"`
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
	Level_5 = "4"

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

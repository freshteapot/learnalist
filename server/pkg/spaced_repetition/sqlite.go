package spaced_repetition

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

const (
	SqlSaveItem   = `INSERT INTO spaced_repetition(uuid, body, user_uuid, when_next, created) values(?, ?, ?, ?, ?)`
	SqlUpdateItem = `UPDATE spaced_repetition SET body=?, when_next=? WHERE user_uuid=? AND uuid=?`
	SqlDeleteItem = `DELETE FROM spaced_repetition WHERE uuid=? AND user_uuid=?`
	SqlGetItem    = `SELECT * FROM spaced_repetition WHERE uuid=? AND user_uuid=?`
	SqlGetAll     = `SELECT body FROM spaced_repetition WHERE user_uuid=? ORDER BY when_next`
	SqlGetNext    = `SELECT * FROM spaced_repetition WHERE user_uuid=? ORDER BY when_next LIMIT 1`
)

type SqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) SpacedRepetitionRepository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) GetNext(userUUID string) (SpacedRepetitionEntry, error) {
	var item SpacedRepetitionEntry

	// TODO might need to update all time stamps to DATETIME as time.Time gets sad when stirng
	err := r.db.Get(&item, SqlGetNext, userUUID)

	if err != nil {
		if err == sql.ErrNoRows {
			return item, ErrNotFound
		}
		return item, err
	}

	return item, nil
}

func (r SqliteRepository) GetEntry(userUUID string, UUID string) (interface{}, error) {
	var body interface{}
	item := SpacedRepetitionEntry{}
	err := r.db.Get(&item, SqlGetItem, UUID, userUUID)

	if err != nil {
		if err == sql.ErrNoRows {
			return body, ErrNotFound
		}

		return body, err
	}

	json.Unmarshal([]byte(item.Body), &body)
	return body, nil
}

func (r SqliteRepository) GetEntries(userUUID string) ([]interface{}, error) {
	items := make([]interface{}, 0)
	dbItems := make([]string, 0)
	// When nothing is found, there is no error.
	err := r.db.Select(&dbItems, SqlGetAll, userUUID)

	if err != nil {
		return items, err
	}

	for _, item := range dbItems {
		var body interface{}
		json.Unmarshal([]byte(item), &body)
		items = append(items, body)
	}

	return items, nil
}

func (r SqliteRepository) SaveEntry(entry SpacedRepetitionEntry) error {
	whenNext := entry.WhenNext.Format(time.RFC3339)
	created := entry.Created.Format(time.RFC3339)
	// TODO Update SQL in production
	_, err := r.db.Exec(SqlSaveItem, entry.UUID, entry.Body, entry.UserUUID, whenNext, created)
	if err != nil {
		fmt.Println(err)
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return ErrSpacedRepetitionEntryExists
		}
		return err
	}
	return nil
}

func (r SqliteRepository) DeleteEntry(userUUID string, UUID string) error {
	_, err := r.db.Exec(SqlDeleteItem, UUID, userUUID)
	return err
}

func (r SqliteRepository) UpdateEntry(entry SpacedRepetitionEntry) error {
	whenNext := entry.WhenNext.Format(time.RFC3339)
	_, err := r.db.Exec(SqlUpdateItem, entry.Body, whenNext, entry.UserUUID, entry.UUID)
	if err != nil {
		return err
	}
	return nil
}

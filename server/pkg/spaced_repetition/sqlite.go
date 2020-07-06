package spaced_repetition

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type SqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) Repository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) GetNext(userUUID string) (interface{}, error) {
	var body interface{}
	item := SpacedRepetitionEntry{}
	// TODO might need to update all time stamps to DATETIME as time.Time gets sad when stirng
	err := r.db.Get(&item, SQL_GET_NEXT, userUUID)

	if err != nil {
		if err == sql.ErrNoRows {
			return body, ErrNotFound
		}
		return body, err
	}

	if !time.Now().UTC().After(item.WhenNext) {
		return body, ErrFoundNotTime
	}

	json.Unmarshal([]byte(item.Body), &body)
	return body, nil
}

func (r SqliteRepository) GetEntry(userUUID string, UUID string) (interface{}, error) {
	var body interface{}
	item := SpacedRepetitionEntry{}
	err := r.db.Get(&item, SQL_GET_ITEM, UUID, userUUID)

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
	err := r.db.Select(&dbItems, SQL_GET_ALL, userUUID)

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
	_, err := r.db.Exec(SQL_SAVE_ITEM, entry.UUID, entry.Body, entry.UserUUID, entry.WhenNext)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return ErrSpacedRepetitionEntryExists
		}
		return err
	}
	return nil
}

func (r SqliteRepository) DeleteEntry(userUUID string, UUID string) error {
	_, err := r.db.Exec(SQL_DELETE_ITEM, UUID, userUUID)
	return err
}

func (r SqliteRepository) UpdateEntry(entry SpacedRepetitionEntry) error {
	_, err := r.db.Exec(SQL_SAVE_ITEM_AUTO_UPDATED, entry.UUID, entry.Body, entry.UserUUID, entry.WhenNext, entry.Body, entry.WhenNext)
	if err != nil {
		return err
	}
	return nil
}

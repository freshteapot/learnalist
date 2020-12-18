package plank

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/jmoiron/sqlx"
)

var (
	SqlGetEntry            = `SELECT body FROM plank WHERE uuid=? AND user_uuid=?`
	SqlSaveEntry           = `INSERT INTO plank(uuid, body, user_uuid, created) values(?, ?, ?, ?)`
	SqlGetHistory          = `SELECT body FROM plank WHERE user_uuid=? ORDER BY created DESC`
	SqlDeleteEntry         = `DELETE FROM plank WHERE uuid=? AND user_uuid=?`
	SqlDeleteEntriesByUser = `DELETE FROM plank WHERE user_uuid=?`
)

type SqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) PlankRepository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) History(userUUID string) ([]openapi.Plank, error) {
	history := make([]openapi.Plank, 0)
	dbItems := make([]string, 0)
	// When nothing is found, there is no error.
	err := r.db.Select(&dbItems, SqlGetHistory, userUUID)

	if err != nil {
		return history, err
	}

	for _, item := range dbItems {
		var body openapi.Plank
		json.Unmarshal([]byte(item), &body)
		history = append(history, body)
	}

	return history, nil
}

func (r SqliteRepository) GetEntry(UUID string, userUUID string) (openapi.Plank, error) {
	var (
		body   string
		record openapi.Plank
	)

	err := r.db.Get(&body, SqlGetEntry, UUID, userUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNotFound
		}
		return record, err
	}

	_ = json.Unmarshal([]byte(body), &record)
	return record, err
}

func (r SqliteRepository) SaveEntry(entry Entry) error {
	created := entry.Created.Format(time.RFC3339)
	b, _ := json.Marshal(entry.Body)
	_, err := r.db.Exec(SqlSaveEntry, entry.UUID, string(b), entry.UserUUID, created)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return ErrEntryExists
		}
		return err
	}
	return nil
}

func (r SqliteRepository) DeleteEntry(UUID string, userUUID string) error {
	_, err := r.db.Exec(SqlDeleteEntry, UUID, userUUID)
	return err
}

func (r SqliteRepository) DeleteEntriesByUser(userUUID string) error {
	_, err := r.db.Exec(SqlDeleteEntriesByUser, userUUID)
	return err
}

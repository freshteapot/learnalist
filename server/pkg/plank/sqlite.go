package plank

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	SqlSaveEntry           = `INSERT INTO plank(uuid, body, user_uuid, created) values(?, ?, ?, ?)`
	SqlGetHistory          = `SELECT body FROM plank WHERE user_uuid=? ORDER BY created DESC`
	SqlDeleteEntry         = `DELETE FROM plank WHERE user_uuid=? and uuid=?`
	SqlDeleteEntriesByUser = `DELETE FROM plank WHERE user_uuid=?`
)

type SqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) Repository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) History(userUUID string) ([]HttpRequestInput, error) {
	history := make([]HttpRequestInput, 0)
	dbItems := make([]string, 0)
	// When nothing is found, there is no error.
	err := r.db.Select(&dbItems, SqlGetHistory, userUUID)

	if err != nil {
		return history, err
	}

	for _, item := range dbItems {
		var body HttpRequestInput
		json.Unmarshal([]byte(item), &body)
		history = append(history, body)
	}

	return history, nil
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

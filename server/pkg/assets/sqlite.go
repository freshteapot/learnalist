package assets

import (
	"github.com/jmoiron/sqlx"
)

var (
	SQL_SAVE_ITEM = `INSERT INTO user_assets(uuid, user_uuid, extension) values(?, ?, ?)`
)

type SqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) Repository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) SaveEntry(entry AssetEntry) error {
	_, err := r.db.Exec(SQL_SAVE_ITEM, entry.UUID, entry.UserUUID, entry.Extension)
	if err != nil {
		return err
	}
	return nil
}

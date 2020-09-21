package assets

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

var (
	SQL_SAVE_ITEM = `INSERT INTO user_assets(uuid, user_uuid, extension) values(?, ?, ?)`
	SQL_GET_ITEM  = `SELECT * FROM user_assets WHERE uuid=?`
)

type SqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) Repository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) GetEntry(UUID string) (AssetEntry, error) {
	item := AssetEntry{}
	err := r.db.Get(&item, SQL_GET_ITEM, UUID)

	if err != nil {
		if err == sql.ErrNoRows {
			return item, ErrNotFound
		}

		return item, err
	}
	return item, nil
}

func (r SqliteRepository) SaveEntry(entry AssetEntry) error {
	_, err := r.db.Exec(SQL_SAVE_ITEM, entry.UUID, entry.UserUUID, entry.Extension)
	if err != nil {
		return err
	}
	return nil
}

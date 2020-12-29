package assets

import (
	"database/sql"

	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/jmoiron/sqlx"
)

var (
	SqlSaveEntry   = `INSERT INTO user_assets(uuid, user_uuid, extension) values(?, ?, ?)`
	SqlGetEntry    = `SELECT * FROM user_assets WHERE uuid=?`
	SqlDeleteEntry = `DELETE FROM user_assets WHERE uuid=? AND user_uuid=?`
)

type SqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) Repository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) DeleteEntry(userUUID string, UUID string) error {
	_, err := r.db.Exec(SqlDeleteEntry, UUID, userUUID)
	return err
}

func (r SqliteRepository) GetEntry(UUID string) (AssetEntry, error) {
	item := AssetEntry{}
	err := r.db.Get(&item, SqlGetEntry, UUID)

	if err != nil {
		if err == sql.ErrNoRows {
			return item, utils.ErrNotFound
		}

		return item, err
	}
	return item, nil
}

func (r SqliteRepository) SaveEntry(entry AssetEntry) error {
	_, err := r.db.Exec(SqlSaveEntry, entry.UUID, entry.UserUUID, entry.Extension)
	if err != nil {
		return err
	}
	return nil
}

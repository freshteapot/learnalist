package dripfeed

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	SqlDeleteByUser                    = `DELETE FROM dripfeed_item WHERE user_uuid=?`
	SqlDeleteByDripfeedUUIDAndUserUUID = `DELETE FROM dripfeed_item WHERE dripfeed_uuid=? AND user_uuid=?`
	SqlDeleteByUserAndSRS              = `DELETE FROM dripfeed_item WHERE user_uuid=? AND srs_uuid=?`
	SqlExists                          = `SELECT 1 FROM dripfeed_item WHERE dripfeed_uuid=?`
	SqlGetNext                         = `
SELECT
	dripfeed_uuid,
	srs_uuid,
	user_uuid,
	alist_uuid,
	body,
	position,
	json_extract(body, '$.kind') AS kind
FROM
	dripfeed_item
WHERE
	dripfeed_uuid=?
ORDER BY
	position
LIMIT 1`
	SqlAddItem = `INSERT OR IGNORE INTO dripfeed_item (dripfeed_uuid, srs_uuid, user_uuid, alist_uuid, body, position) VALUES (?, ?, ?, ?, ?, ?)`
)

type sqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) DripfeedRepository {
	return sqliteRepository{
		db: db,
	}
}

func (r sqliteRepository) DeleteByUser(userUUID string) error {
	_, err := r.db.Exec(SqlDeleteByUser, userUUID)
	return err
}

func (r sqliteRepository) Exists(dripfeedUUID string) (bool, error) {
	var id int
	err := r.db.Get(&id, SqlExists, dripfeedUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		// Set to true, incase people are not listening for err
		return true, err
	}

	if id != 1 {
		return false, nil
	}
	return true, nil
}

func (r sqliteRepository) GetNext(dripfeedUUID string) (RepoItem, error) {
	type dbItem struct {
		SrsUUID      string `db:"srs_uuid"`
		SrsKind      string `db:"kind"`
		SrsBody      []byte `db:"body"`
		Position     int    `db:"position"`
		DripfeedUUID string `db:"dripfeed_uuid"`
		UserUUID     string `db:"user_uuid"`
		AlistUUID    string `db:"alist_uuid"`
	}

	item := dbItem{}
	// json_extract(body, '$.display_name'
	err := r.db.Get(&item, SqlGetNext, dripfeedUUID)
	if err != nil {
		panic(err)
	}

	return RepoItem{
		SrsUUID:      item.SrsUUID,
		SrsKind:      item.SrsKind,
		SrsBody:      item.SrsBody,
		Position:     item.Position,
		DripfeedUUID: item.DripfeedUUID,
		UserUUID:     item.UserUUID,
		AlistUUID:    item.AlistUUID,
	}, nil
}

func (r sqliteRepository) AddAll(dripfeedUUID string, userUUID string, alistUUID string, items []interface{}) error {
	for index, item := range items {
		body := item.(string)
		var srs SpacedRepetitionUUID
		json.Unmarshal([]byte(body), &srs)
		fmt.Println("srsUUID", srs.UUID)
		_, err := r.db.Exec(
			SqlAddItem,
			dripfeedUUID,
			srs.UUID,
			userUUID,
			alistUUID,
			body,
			index)
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func (r sqliteRepository) DeleteByPosition(dripfeedUUID string, position int) error {
	return errors.New("TODO")
}

// This could be all
func (r sqliteRepository) DeleteBySpacedRepetitionUUID(dripfeedUUID string, srsUUID string) error {
	return errors.New("TODO")
}

func (r sqliteRepository) DeleteAllByUserUUIDAndSpacedRepetitionUUID(userUUID string, srsUUID string) error {
	_, err := r.db.Exec(SqlDeleteByUserAndSRS, userUUID, srsUUID)
	return err
}

func (r sqliteRepository) DeleteByUUIDAndUserUUID(dripfeedUUID string, userUUID string) error {
	_, err := r.db.Exec(SqlDeleteByDripfeedUUIDAndUserUUID, dripfeedUUID, userUUID)
	return err
}

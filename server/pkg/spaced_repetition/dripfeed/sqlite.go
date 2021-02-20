package dripfeed

import (
	"database/sql"
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/jmoiron/sqlx"
)

const (
	SqlGetDripfeedInfo                     = `SELECT dripfeed_uuid, user_uuid, alist_uuid FROM dripfeed_info WHERE dripfeed_uuid=?`
	SqlSaveDripfeedInfo                    = `INSERT OR IGNORE INTO dripfeed_info (dripfeed_uuid, user_uuid, alist_uuid) VALUES (?, ?, ?)`
	SqlDeleteInfoByDripfeedUUIDAndUserUUID = `DELETE FROM dripfeed_info WHERE dripfeed_uuid=? AND user_uuid=?`
	SqlDeleteInfoByUser                    = `DELETE FROM dripfeed_info WHERE user_uuid=?`

	SqlDeleteDripfeedItemByUser                    = `DELETE FROM dripfeed_item WHERE user_uuid=?`
	SqlDeleteDripfeedItemByDripfeedUUIDAndUserUUID = `DELETE FROM dripfeed_item WHERE dripfeed_uuid=? AND user_uuid=?`
	SqlDeleteDripfeedItemByUserAndSRS              = `DELETE FROM dripfeed_item WHERE user_uuid=? AND srs_uuid=?`
	SqlDripfeedItemExists                          = `SELECT 1 FROM dripfeed_item WHERE dripfeed_uuid=?`
	SqlDripfeedItemGetNext                         = `
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
	SqlDripfeedItemAddItem = `INSERT OR IGNORE INTO dripfeed_item (dripfeed_uuid, srs_uuid, user_uuid, alist_uuid, body, position) VALUES (?, ?, ?, ?, ?, ?)`
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
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(SqlDeleteDripfeedItemByUser, userUUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(SqlDeleteInfoByUser, userUUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// Exists
// True false or error
func (r sqliteRepository) Exists(dripfeedUUID string) (bool, error) {
	var id int
	err := r.db.Get(&id, SqlDripfeedItemExists, dripfeedUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		// Set to true, incase people are not listening for err
		return true, err
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
	err := r.db.Get(&item, SqlDripfeedItemGetNext, dripfeedUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return RepoItem{}, utils.ErrNotFound
		}
		return RepoItem{}, err
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

func (r sqliteRepository) AddAll(dripfeedUUID string, userUUID string, alistUUID string, items []string) error {
	err := r.SaveInfo(openapi.SpacedRepetitionOvertimeInfo{
		DripfeedUuid: dripfeedUUID,
		UserUuid:     userUUID,
		AlistUuid:    alistUUID,
	})
	if err != nil {
		panic(err)
	}

	for index, body := range items {
		var srs SpacedRepetitionUUID
		json.Unmarshal([]byte(body), &srs)
		_, err = r.db.Exec(
			SqlDripfeedItemAddItem,
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

func (r sqliteRepository) DeleteAllByUserUUIDAndSpacedRepetitionUUID(userUUID string, srsUUID string) error {
	_, err := r.db.Exec(SqlDeleteDripfeedItemByUserAndSRS, userUUID, srsUUID)
	return err
}

func (r sqliteRepository) DeleteByUUIDAndUserUUID(dripfeedUUID string, userUUID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(SqlDeleteDripfeedItemByDripfeedUUIDAndUserUUID, dripfeedUUID, userUUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(SqlDeleteInfoByDripfeedUUIDAndUserUUID, dripfeedUUID, userUUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (r sqliteRepository) GetInfo(dripfeedUUID string) (openapi.SpacedRepetitionOvertimeInfo, error) {

	type dbItem struct {
		DripfeedUUID string `db:"dripfeed_uuid"`
		UserUUID     string `db:"user_uuid"`
		AlistUUID    string `db:"alist_uuid"`
	}

	response := openapi.SpacedRepetitionOvertimeInfo{}
	item := dbItem{}
	err := r.db.Get(&item, SqlGetDripfeedInfo, dripfeedUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response, utils.ErrNotFound
		}
		return response, err
	}

	response.AlistUuid = item.AlistUUID
	response.DripfeedUuid = item.DripfeedUUID
	response.UserUuid = item.UserUUID
	return response, nil
}

func (r sqliteRepository) SaveInfo(input openapi.SpacedRepetitionOvertimeInfo) error {
	_, err := r.db.Exec(
		SqlSaveDripfeedInfo,
		input.DripfeedUuid,
		input.UserUuid,
		input.AlistUuid)
	return err
}

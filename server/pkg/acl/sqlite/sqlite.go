package sqlite

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type DatabaseAcl struct {
	ExtUUID  string `db:"ext_uuid"`
	UserUUID string `db:"user_uuid"`
	Access   string `db:"access"`
}

var noUserUUUID = "nouser"
var insertAccess = `
INSERT INTO acl_simple (ext_uuid, user_uuid, access)
VALUES(:ext_uuid, :user_uuid, :access);`

var deleteViaAccess = `DELETE FROM acl_simple WHERE access = ?`
var deleteViaAlistUUID = `DELETE FROM acl_simple WHERE ext_uuid = ?`
var selectAccessDirect = `SELECT access FROM acl_simple WHERE access = ?`
var selectAccessFilter = `SELECT access FROM acl_simple WHERE access LIKE ?`

type Sqlite struct {
	db *sqlx.DB
}

func NewAcl(db *sqlx.DB) *Sqlite {
	return &Sqlite{
		db: db,
	}
}

func (store *Sqlite) DeleteByExtUUID(extUUID string) error {
	_, err := store.db.Exec(deleteViaAlistUUID, extUUID)
	return err
}

func (store *Sqlite) accessExsits(access string) (bool, error) {
	var allow string
	err := store.db.Get(&allow, selectAccessDirect, access)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, nil
		}
		// TODO this would need logging
		return false, err
	}
	return true, nil
}

func (store *Sqlite) insert(extUUID string, userUUID string, access string) error {
	data := &DatabaseAcl{
		ExtUUID:  extUUID,
		UserUUID: userUUID,
		Access:   access,
	}

	_, err := store.db.NamedExec(insertAccess, data)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: acl_simple.ext_uuid, acl_simple.user_uuid, acl_simple.access" {
			return nil
		}
		return err
	}
	return nil
}

func (store *Sqlite) deleteViaAccess(access string) error {
	_, err := store.db.Exec(deleteViaAccess, access)
	// TODO handle err
	return err
}

func insertTX(tx *sqlx.Tx, extUUID string, userUUID string, access string) (sql.Result, error) {
	data := &DatabaseAcl{
		ExtUUID:  extUUID,
		UserUUID: userUUID,
		Access:   access,
	}
	return tx.NamedExec(insertAccess, data)
}

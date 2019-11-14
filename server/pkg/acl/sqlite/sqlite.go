package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/jmoiron/sqlx"
)

type DatabaseAcl struct {
	AlistUUID string `db:"alist_uuid"`
	UserUUID  string `db:"user_uuid"`
	Access    string `db:"access"`
}

var noUserUUUID = "nouser"
var insertAccess = `
INSERT INTO acl_simple (alist_uuid, user_uuid, access)
VALUES(:alist_uuid, :user_uuid, :access);`

var deleteViaAccess = `DELETE FROM acl_simple WHERE access = ?`
var deleteViaAlistUUID = `DELETE FROM acl_simple WHERE alist_uuid = ?`
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

func (store *Sqlite) GrantUserListWriteAccess(alistUUID string, userUUID string) error {
	access := fmt.Sprintf(keys.ListWriteAccessForUser, alistUUID, userUUID)
	return store.insert(alistUUID, userUUID, access)
}

func (store *Sqlite) RevokeUserListWriteAccess(alistUUID string, userUUID string) error {
	access := fmt.Sprintf(keys.ListWriteAccessForUser, alistUUID, userUUID)
	return store.deleteViaAccess(access)
}

func (store *Sqlite) GrantUserListReadAccess(alistUUID string, userUUID string) error {
	access := fmt.Sprintf(keys.ListReadAccessForUser, alistUUID, userUUID)
	return store.insert(alistUUID, userUUID, access)
}
func (store *Sqlite) RevokeUserListReadAccess(alistUUID string, userUUID string) error {
	access := fmt.Sprintf(keys.ListReadAccessForUser, alistUUID, userUUID)
	return store.deleteViaAccess(access)
}

func (store *Sqlite) ShareListWithPublic(alistUUID string) error {
	accessPrivate := fmt.Sprintf(keys.ListSharePrivate, alistUUID)
	accessFriends := fmt.Sprintf(keys.ListShareFriends, alistUUID)
	accessPublic := fmt.Sprintf(keys.ListSharePublic, alistUUID)

	tx, err := store.db.Beginx()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(deleteViaAccess, accessPrivate)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(deleteViaAccess, accessFriends)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = insertTX(tx, alistUUID, noUserUUUID, accessPublic)
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

func (store *Sqlite) MakeListPrivate(alistUUID string, userUUID string) error {
	read := fmt.Sprintf(keys.ListReadAccessForUser, alistUUID, userUUID)
	owner := fmt.Sprintf(keys.ListOwnerAccessForUser, alistUUID, userUUID)
	share := fmt.Sprintf(keys.ListSharePrivate, alistUUID)

	tx, err := store.db.Beginx()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(deleteViaAlistUUID, alistUUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = insertTX(tx, alistUUID, noUserUUUID, read)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = insertTX(tx, alistUUID, noUserUUUID, owner)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = insertTX(tx, alistUUID, noUserUUUID, share)
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

func (store *Sqlite) DeleteList(alistUUID string) error {
	_, err := store.db.Exec(deleteViaAlistUUID, alistUUID)
	return err
}

func (store *Sqlite) ShareListWithFriends(alistUUID string) error {
	accessPrivate := fmt.Sprintf(keys.ListSharePrivate, alistUUID)
	accessFriends := fmt.Sprintf(keys.ListShareFriends, alistUUID)
	accessPublic := fmt.Sprintf(keys.ListSharePublic, alistUUID)

	tx, err := store.db.Beginx()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(deleteViaAccess, accessPrivate)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(deleteViaAccess, accessPublic)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = insertTX(tx, alistUUID, noUserUUUID, accessFriends)
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

func (store *Sqlite) IsListPublic(alistUUID string) (bool, error) {
	access := fmt.Sprintf(keys.ListSharePublic, alistUUID)
	return store.accessExsits(access)
}

func (store *Sqlite) IsListPrivate(alistUUID string) (bool, error) {
	access := fmt.Sprintf(keys.ListSharePrivate, alistUUID)
	return store.accessExsits(access)
}

func (store *Sqlite) IsListAvailableToFriends(alistUUID string) (bool, error) {
	access := fmt.Sprintf(keys.ListShareFriends, alistUUID)
	return store.accessExsits(access)
}

func (store *Sqlite) ListIsSharedWith(alistUUID string) (string, error) {
	var with string
	filter := fmt.Sprintf(keys.FilterListShare, alistUUID)
	err := store.db.Get(&with, selectAccessFilter, filter)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return with, nil
		}
		// TODO this would need logging
		return with, err
	}

	parts := strings.Split(with, ":")

	switch parts[4] {
	case keys.SharedWithPublic:
		return keys.SharedWithPublic, nil
	case keys.NotShared:
		return keys.NotShared, nil
	case keys.SharedWithFriends:
		return keys.SharedWithFriends, nil
	default:
		return "", errors.New("something is saved that is not supported")
	}
}

func (store *Sqlite) HasUserListReadAccess(alistUUID string, userUUID string) (bool, error) {
	access := fmt.Sprintf(keys.ListReadAccessForUser, alistUUID, userUUID)
	allow, err := store.accessExsits(access)
	if err != nil {
		return false, err
	}
	if !allow {
		return store.IsListPublic(alistUUID)
	}
	return allow, nil
}

func (store *Sqlite) HasUserListWriteAccess(alistUUID string, userUUID string) (bool, error) {
	access := fmt.Sprintf(keys.ListWriteAccessForUser, alistUUID, userUUID)
	return store.accessExsits(access)
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

func (store *Sqlite) insert(alistUUID string, userUUID string, access string) error {
	data := &DatabaseAcl{
		AlistUUID: alistUUID,
		UserUUID:  userUUID,
		Access:    access,
	}

	_, err := store.db.NamedExec(insertAccess, data)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: acl_simple.alist_uuid, acl_simple.user_uuid, acl_simple.access" {
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

func insertTX(tx *sqlx.Tx, alistUUID string, userUUID string, access string) (sql.Result, error) {
	data := &DatabaseAcl{
		AlistUUID: alistUUID,
		UserUUID:  userUUID,
		Access:    access,
	}
	return tx.NamedExec(insertAccess, data)
}

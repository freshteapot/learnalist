package sqlite

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
)

func (store *Sqlite) ShareKindWithPublic(kind string, extUUID string) error {
	accessPrivate := fmt.Sprintf(keys.KindSharePrivate, kind, extUUID)
	accessFriends := fmt.Sprintf(keys.KindShareFriends, kind, extUUID)
	accessPublic := fmt.Sprintf(keys.KindSharePublic, kind, extUUID)

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

	_, err = insertTX(tx, extUUID, noUserUUUID, accessPublic)
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

func (store *Sqlite) MakeKindPrivate(kind string, extUUID string, userUUID string) error {
	read := fmt.Sprintf(keys.KindReadAccessForUser, kind, extUUID, userUUID)
	owner := fmt.Sprintf(keys.KindOwnerAccessForUser, kind, extUUID, userUUID)
	share := fmt.Sprintf(keys.KindSharePrivate, kind, extUUID)

	tx, err := store.db.Beginx()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(deleteViaAlistUUID, extUUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = insertTX(tx, extUUID, noUserUUUID, read)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = insertTX(tx, extUUID, noUserUUUID, owner)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = insertTX(tx, extUUID, noUserUUUID, share)
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

func (store *Sqlite) ShareKindWithFriends(kind string, extUUID string) error {
	accessPrivate := fmt.Sprintf(keys.KindSharePrivate, kind, extUUID)
	accessFriends := fmt.Sprintf(keys.KindShareFriends, kind, extUUID)
	accessPublic := fmt.Sprintf(keys.KindSharePublic, kind, extUUID)

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

	_, err = insertTX(tx, extUUID, noUserUUUID, accessFriends)
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

func (store *Sqlite) IsKindPublic(kind string, extUUID string) (bool, error) {
	access := fmt.Sprintf(keys.KindSharePublic, kind, extUUID)
	return store.accessExsits(access)
}

func (store *Sqlite) IsKindPrivate(kind string, extUUID string) (bool, error) {
	access := fmt.Sprintf(keys.KindSharePrivate, kind, extUUID)
	return store.accessExsits(access)
}

func (store *Sqlite) IsKindAvailableToFriends(kind string, extUUID string) (bool, error) {
	access := fmt.Sprintf(keys.KindShareFriends, kind, extUUID)
	return store.accessExsits(access)
}

func (store *Sqlite) HasUserKindReadAccess(kind string, extUUID string, userUUID string) (bool, error) {
	access := fmt.Sprintf(keys.KindReadAccessForUser, kind, extUUID, userUUID)
	allow, err := store.accessExsits(access)
	if err != nil {
		return false, err
	}

	if !allow {
		return store.IsKindPublic(kind, extUUID)
	}
	return allow, nil
}

func (store *Sqlite) HasUserKindWriteAccess(kind string, extUUID string, userUUID string) (bool, error) {
	access := fmt.Sprintf(keys.KindWriteAccessForUser, kind, extUUID, userUUID)
	return store.accessExsits(access)
}

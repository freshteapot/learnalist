package sqlite

import (
	"errors"
	"fmt"
	"strings"

	"github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
)

func (store *Sqlite) GrantUserListWriteAccess(alistUUID string, userUUID string) error {
	return store.GrantUserKindWriteAccess("list", alistUUID, userUUID)
}

func (store *Sqlite) RevokeUserListWriteAccess(alistUUID string, userUUID string) error {
	return store.RevokeUserKindWriteAccess("list", alistUUID, userUUID)
}

func (store *Sqlite) GrantUserListReadAccess(alistUUID string, userUUID string) error {
	return store.GrantUserKindReadAccess("list", alistUUID, userUUID)
}
func (store *Sqlite) RevokeUserListReadAccess(alistUUID string, userUUID string) error {
	return store.RevokeUserKindReadAccess("list", alistUUID, userUUID)
}

func (store *Sqlite) ShareListWithPublic(alistUUID string) error {
	return store.ShareKindWithPublic("list", alistUUID)
}

func (store *Sqlite) MakeListPrivate(alistUUID string, userUUID string) error {
	return store.MakeKindPrivate("list", alistUUID, userUUID)
}

func (store *Sqlite) DeleteList(alistUUID string) error {
	return store.DeleteByExtUUID(alistUUID)
}

func (store *Sqlite) ShareListWithFriends(alistUUID string) error {
	return store.ShareKindWithFriends("list", alistUUID)
}

func (store *Sqlite) IsListPublic(alistUUID string) (bool, error) {
	return store.IsKindPublic("list", alistUUID)
}

func (store *Sqlite) IsListPrivate(alistUUID string) (bool, error) {
	return store.IsKindPrivate("list", alistUUID)
}

func (store *Sqlite) IsListAvailableToFriends(alistUUID string) (bool, error) {
	return store.IsKindAvailableToFriends("list", alistUUID)
}

func (store *Sqlite) HasUserListReadAccess(alistUUID string, userUUID string) (bool, error) {
	return store.HasUserKindReadAccess("list", alistUUID, userUUID)
}

func (store *Sqlite) HasUserListWriteAccess(alistUUID string, userUUID string) (bool, error) {
	return store.HasUserKindWriteAccess("list", alistUUID, userUUID)
}

func (store *Sqlite) HasUserPublicListWriteAccess(userUUID string) (bool, error) {
	return store.HasUserKindWriteAccess("list", keys.SharedWithPublic, userUUID)
}

// Grant a user access public write access
func (store *Sqlite) GrantUserPublicListWriteAccess(userUUID string) error {
	return store.GrantUserKindWriteAccess("list", keys.SharedWithPublic, userUUID)
}

// Revoke a users right to write public lists
func (store *Sqlite) RevokeUserPublicListWriteAccess(userUUID string) error {
	return store.RevokeUserKindWriteAccess("list", keys.SharedWithPublic, userUUID)
}

// TODO what is this trickery.
// TODO nothing uses this, I wonder what the thought was
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

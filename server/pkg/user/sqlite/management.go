package sqlite

import (
	"errors"

	"github.com/jmoiron/sqlx"
)

type sqliteManagement struct {
	db *sqlx.DB
}

func NewSqliteManagementStorage(db *sqlx.DB) sqliteManagement {
	return sqliteManagement{db: db}
}

// FindUserUUID Find the user uuid based on the search string
func (m sqliteManagement) FindUserUUID(search string) ([]string, error) {
	db := m.db
	query := `
SELECT
	uuid as user_uuid
FROM
	user
WHERE
	username=?
UNION
SELECT
	user_uuid
FROM
	user_from_idp
WHERE
	kind="email"
AND
	identifier=?
UNION
SELECT
	user_uuid
FROM
	user_from_idp
WHERE
	user_uuid=?
UNION
SELECT
	uuid as user_uuid
FROM
	user
WHERE
	uuid=?`

	userUUIDs := make([]string, 0)
	err := db.Select(&userUUIDs, query, search, search, search, search)

	if len(userUUIDs) > 1 {
		return userUUIDs, errors.New("Too many userUUID found")
	}
	return userUUIDs, err
}

func (m sqliteManagement) GetLists(userUUID string) ([]string, error) {
	lists := make([]string, 0)
	query := `
SELECT
	uuid
FROM
	alist_kv
WHERE
	user_uuid=?`

	err := m.db.Select(&lists, query, userUUID)

	return lists, err
}

func (m sqliteManagement) DeleteUser(userUUID string) error {
	db := m.db
	queries := []string{
		"DELETE FROM user_labels WHERE user_uuid=?",
		"DELETE FROM user WHERE uuid=?",
		"DELETE FROM alist_labels WHERE user_uuid=?",
		"DELETE FROM acl_simple WHERE user_uuid=?",
		"DELETE FROM oauth2_token_info WHERE user_uuid=?",
		"DELETE FROM user_sessions WHERE user_uuid=?",
		"DELETE FROM user_from_idp WHERE user_uuid=?"}

	tx, err := db.Beginx()
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, query := range queries {
		_, err = tx.Exec(query, userUUID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (m sqliteManagement) DeleteList(listUUID string) error {
	db := m.db
	queries := []string{
		"DELETE FROM alist_labels WHERE alist_uuid=?",
		"DELETE FROM acl_simple WHERE ext_uuid=?",
		"DELETE FROM alist_kv WHERE uuid=?"}

	tx, err := db.Beginx()
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, query := range queries {
		_, err = tx.Exec(query, listUUID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

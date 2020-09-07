package sqlite

import (
	"github.com/freshteapot/learnalist-api/server/api/user"

	"github.com/jmoiron/sqlx"
)

type store struct {
	db *sqlx.DB
}

const (
	SqlInsertList              = `INSERT INTO alist_kv(uuid, list_type, body, user_uuid) values(?, ?, ?, ?)`
	SqlUpdateList              = `UPDATE alist_kv SET list_type=?, body=?, user_uuid=? WHERE uuid=?`
	SqlGetItemByUUID           = `SELECT uuid, body, user_uuid, list_type FROM alist_kv WHERE uuid = ?`
	SqlDeleteItemByUserAndUUID = `
DELETE
FROM
	alist_kv
WHERE
	uuid=?
AND
	user_uuid=?
`
	SqlGetPublicLists = `
SELECT
	uuid,
	title
FROM (
SELECT
	json_extract(body, '$.info.title') AS title,
	IFNULL(json_extract(body, '$.info.shared_with'), "private") AS shared_with,
	uuid
FROM
	alist_kv
) as temp
WHERE
	shared_with="public";
	`
)

func NewUser(db *sqlx.DB) user.DatastoreUsers {
	return &store{
		db: db,
	}
}

func (store *store) UserExists(userUUID string) bool {
	var id int
	query := `
SELECT 1 FROM user WHERE uuid=?
UNION
SELECT 1 FROM user_from_idp WHERE user_uuid=?
`
	store.db.Get(&id, query, userUUID, userUUID)
	if id != 1 {
		return false
	}
	return true
}

package sqlite

import (
	"database/sql"

	"github.com/freshteapot/learnalist-api/server/pkg/user"
	guuid "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type DatabaseUserFromIDP struct {
	UserUUID   string `db:"user_uuid"`
	IDP        string `db:"idp"`
	Identifier string `db:"identifier"`
	Kind       string `db:"kind"`
	Info       string `db:"info"`
	Created    int64  `db:"created"`
}

type UserFromIDP struct {
	db *sqlx.DB
}

const (
	UserFromIDPInsertEntry  = `INSERT INTO user_from_idp (user_uuid, idp, identifier, kind, info) VALUES (?, ?, ?, ?, ?)`
	UserFromIDPFindUserUUID = `
SELECT
	user_uuid, idp, identifier, kind, info, created
FROM
	user_from_idp
WHERE
	idp=?
AND
	kind="email"
AND
	identifier=?`

	UserFromIDPSelectByUserUUID = `
SELECT
	user_uuid, idp, identifier, kind, info, created
FROM
	user_from_idp
WHERE
	user_uuid=?`
)

func NewUserFromIDP(db *sqlx.DB) *UserFromIDP {
	return &UserFromIDP{
		db: db,
	}
}

func (store *UserFromIDP) Register(idp string, identifier string, info []byte) (userUUID string, err error) {
	id := guuid.New()
	userUUID = id.String()
	_, err = store.db.Exec(UserFromIDPInsertEntry, userUUID, idp, identifier, "email", string(info))
	return userUUID, err
}

func (store *UserFromIDP) Lookup(idp string, identifier string) (userUUID string, err error) {
	var item DatabaseUserFromIDP
	err = store.db.Get(&item, UserFromIDPFindUserUUID, idp, identifier)
	if err == sql.ErrNoRows {
		err = user.NotFound
	}
	return item.UserUUID, err
}

package sqlite

import (
	"database/sql"

	"github.com/freshteapot/learnalist-api/server/pkg/user"
	guuid "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type DatabaseUser struct {
	Uuid     string `db:"uuid"`
	Username string `db:"username"`
	Hash     string `db:"hash"`
}

type UserWithUsernameAndPassword struct {
	db *sqlx.DB
}

const (
	UserWithUsernameAndPasswordInsertEntry = `INSERT INTO user (uuid, username, hash) VALUES (?, ?, ?)`

	UserWithUsernameAndPasswordSelectUserUUIDByHash = `
SELECT
	uuid
FROM
	user
WHERE
	username=?
AND
	hash=?`
)

func NewUserWithUsernameAndPassword(db *sqlx.DB) *UserWithUsernameAndPassword {
	return &UserWithUsernameAndPassword{
		db: db,
	}
}

func (store *UserWithUsernameAndPassword) Register(username string, hash string) (info user.UserInfoFromUsernameAndPassword, err error) {
	info.Username = username
	info.Hash = hash

	// Does the user already exist
	userUUID, err := store.Lookup(username, hash)
	if err == nil {
		info.UserUUID = userUUID
		return info, nil
	}

	if err != nil {
		if err != sql.ErrNoRows {
			return info, err
		}
	}

	// Assume the user doesnt exist and try and insert them.
	id := guuid.New()
	user := &DatabaseUser{
		Uuid:     id.String(),
		Hash:     hash,
		Username: username,
	}

	_, err = store.db.Exec(UserWithUsernameAndPasswordInsertEntry, user.Uuid, user.Username, user.Hash)
	if err.Error() != "UNIQUE constraint failed: user.username" {
		return info, err
	}

	info.UserUUID = id.String()
	return info, err
}

func (store *UserWithUsernameAndPassword) Lookup(username string, hash string) (userUUID string, err error) {
	err = store.db.Get(&userUUID, UserWithUsernameAndPasswordSelectUserUUIDByHash, username, hash)
	return userUUID, err
}

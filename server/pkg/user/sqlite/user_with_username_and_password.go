package sqlite

import (
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

// Register a new user with username and hashed password
func (store *UserWithUsernameAndPassword) Register(username string, hash string) (info user.UserInfoFromUsernameAndPassword, err error) {
	id := guuid.New()
	info.Username = username
	info.Hash = hash
	info.UserUUID = id.String()

	_, err = store.db.Exec(UserWithUsernameAndPasswordInsertEntry, info.UserUUID, info.Username, info.Hash)
	return info, err
}

func (store *UserWithUsernameAndPassword) Lookup(username string, hash string) (userUUID string, err error) {
	err = store.db.Get(&userUUID, UserWithUsernameAndPasswordSelectUserUUIDByHash, username, hash)
	return userUUID, err
}

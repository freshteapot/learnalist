package sqlite

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/jmoiron/sqlx"
)

type UserFromIDP struct {
	db *sqlx.DB
}

const (
	UserFromIDPInsertEntry      = `INSERT INTO user_sessions (challenge) VALUES (?)`
	UserFromIDPUpdateEntry      = `UPDATE user_sessions SET token=? AND user_uuid=? WHERE challenge=? AND token="none"`
	UserFromIDPDeleteByUserUUID = `DELETE FROM user_sessions WHERE user_uuid = ?`
)

func NewUserFromIDP(db *sqlx.DB) *UserFromIDP {
	return &UserFromIDP{
		db: db,
	}
}

func (store *UserFromIDP) Register(from string, kind string, identifier string, info []byte) (user.UserUUID, error) {
	var userUUID user.UserUUID

	userUUID = "fake-user-123"
	return userUUID, errors.New("todo")
}

func (store *UserFromIDP) Lookup(from string, kind string, identifier string) (user.UserUUID, error) {
	return "", errors.New("todo")
}

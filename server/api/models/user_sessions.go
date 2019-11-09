package models

import (
	"errors"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
)

const SessionsWithoutUser = "na"

type DatabaseUserSession struct {
	Token    string    `db:"token"`
	UserUUID string    `db:"user_uuid"`
	Created  time.Time `db:"created"`
}

// InsertNewSession, create a random token
func (dal *DAL) InsertNewSession() (string, error) {
	token := uuid.GetUUID("user.session")
	session := &DatabaseUserSession{
		Token:    token,
		UserUUID: SessionsWithoutUser,
		Created:  time.Now().UTC(),
	}

	query := `INSERT INTO user_sessions(token, user_uuid, created) VALUES (?, ?, ?);`

	_, err := dal.Db.Exec(query, session.Token, session.UserUUID, session.Created)
	return session.Token, err
}

// GetSessionByToken look up the session vi the token
func (dal *DAL) GetSessionByToken(token string) (DatabaseUserSession, error) {
	var session DatabaseUserSession
	err := dal.Db.Get(&session, `SELECT * FROM user_sessions WHERE token=?`, token)
	return session, err
}

// UpdateSession link the userUUID to the token, to make the sesion real
func (dal *DAL) UpdateSession(userUUID string, token string) error {
	query := `UPDATE user_sessions SET user_uuid=? WHERE token=?;`
	result, err := dal.Db.Exec(query, userUUID, token)
	if err != nil {
		return err
	}

	total, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if total == 0 {
		return errors.New("token not found, failing to link the user.")
	}
	return nil
}

package sqlite

import (
	"database/sql"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	guuid "github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserSession struct {
	db *sqlx.DB
}

const (
	UserSessionInsertEntry              = `INSERT INTO user_sessions (challenge) VALUES (?)`
	UserSessionInsertFullRecord         = `INSERT INTO user_sessions (challenge, token, user_uuid, created) VALUES (?, ?, ?, ?)`
	UserSessionUpdateEntry              = `UPDATE user_sessions SET token=?, user_uuid=? WHERE challenge=? AND token="none"`
	UserSessionDeleteByUserUUID         = `DELETE FROM user_sessions WHERE user_uuid=?`
	UserSessionDeleteByUserUUIDAndToken = `DELETE FROM user_sessions WHERE user_uuid=? AND token=?`
	UserSessionDeleteUnActiveChallenges = `
DELETE FROM
	user_sessions
WHERE
	token = "none"
AND
	user_uuid = "none"
AND
	datetime(created, 'unixepoch') < datetime(
		datetime(strftime('%s','now'),'unixepoch'),
		'-12 hour');
`
	UserSessionSelectByToken = `
SELECT
	challenge, token, user_uuid, created
FROM
	user_sessions
WHERE token = ?`

	UserSessionSelectUserUUIDByToken = `
SELECT
	user_uuid
FROM
	user_sessions
WHERE token = ?`

	UserSessionSelectChallengeIsValid = `SELECT 1 FROM user_sessions WHERE challenge=? AND token="none"`
)

func NewUserSession(db *sqlx.DB) *UserSession {
	return &UserSession{
		db: db,
	}
}

func (store *UserSession) NewSession(userUUID string) (session user.UserSession, err error) {
	token := guuid.New()
	challenge := guuid.New()
	when := time.Now().UTC()

	session.UserUUID = userUUID
	session.Token = token.String()
	session.Challenge = challenge.String()
	session.Created = when

	_, err = store.db.Exec(UserSessionInsertFullRecord, session.Challenge, session.Token, session.UserUUID, when.Unix())
	return session, err
}

func (store *UserSession) CreateWithChallenge() (string, error) {
	id := guuid.New()
	_, err := store.db.Exec(UserSessionInsertEntry, id.String())
	return id.String(), err
}

func (store *UserSession) Activate(session user.UserSession) error {
	info, err := store.db.Exec(UserSessionUpdateEntry, &session.Token, session.UserUUID, session.Challenge)
	if err != nil {
		return err
	}

	total, err := info.RowsAffected()
	if err != nil {
		return err
	}

	if total != 1 {
		return i18n.ErrorUserSessionActivate
	}

	return nil
}

func (store *UserSession) GetUserUUIDByToken(token string) (userUUID string, err error) {
	err = store.db.Get(&userUUID, UserSessionSelectUserUUIDByToken, token)
	return userUUID, err
}

func (store *UserSession) IsChallengeValid(challenge string) (bool, error) {
	var id int
	err := store.db.Get(&id, UserSessionSelectChallengeIsValid, challenge)
	if err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}
	}

	if id != 1 {
		return false, nil
	}
	return true, nil
}

func (store *UserSession) RemoveSessionsForUser(userUUID string) error {
	_, err := store.db.Exec(UserSessionDeleteByUserUUID, userUUID)
	return err
}

func (store *UserSession) RemoveExpiredChallenges() error {
	_, err := store.db.Exec(UserSessionDeleteUnActiveChallenges)
	return err
}

func (store *UserSession) RemoveSessionForUser(userUUID string, token string) error {
	_, err := store.db.Exec(UserSessionDeleteByUserUUIDAndToken, userUUID, token)
	return err
}

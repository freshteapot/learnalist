package sqlite

import (
	"database/sql"

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
	UserSessionUpdateEntry              = `UPDATE user_sessions SET token=?, user_uuid=? WHERE challenge=? AND token="none"`
	UserSessionDeleteByUserUUID         = `DELETE FROM user_sessions WHERE user_uuid = ?`
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

	UserSessionSelectChallengeIsValid = `SELECT 1 FROM user_sessions WHERE challenge=? AND token="none"`
)

func NewUserSession(db *sqlx.DB) *UserSession {
	return &UserSession{
		db: db,
	}
}

func (store *UserSession) Create() (string, error) {
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

func (store *UserSession) Get(token string) (user.UserSession, error) {
	var session user.UserSession

	err := store.db.Get(&session, UserSessionSelectByToken, token)
	return session, err
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

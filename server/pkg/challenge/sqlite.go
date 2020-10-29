package challenge

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	SqlGetEntry    = `SELECT * FROM challenge WHERE uuid=?`
	SqlSaveEntry   = `INSERT INTO challenge(uuid, user_uuid, body) values(?, ?, ?)`
	SqlDeleteEntry = `DELETE FROM challenge WHERE uuid=?`
	//SqlDeleteEntriesByUser = `DELETE FROM challenge WHERE user_uuid=?`
	SqlAddRecord    = `INSERT INTO challenge_record(uuid, ext_uuid, user_uuid) values(?, ?, ?)`
	SqlDeleteRecord = `DELETE FROM challenge_record WHERE ext_uuid=? AND user_uuid=?`
	//SqlDeleteRecords       = `DELETE FROM challenge_record WHERE uuid=?`
	// Tightly couple the planks with the challenges for now.
	SqlPlankRecords = `
SELECT
	c.uuid, c.user_uuid, p.body
FROM
	plank AS p
INNER JOIN
	challenge_record AS c
WHERE
	c.uuid = ?
AND
	p.uuid = c.ext_uuid
ORDER BY
	p.created
DESC
`
	SqlGetChallengesByUser = `
SELECT c.*
FROM challenge AS c
INNER JOIN
(
SELECT REPLACE(
	REPLACE(access,"api:challenge:", ""),
	":write:%s", ""
	) AS uuid
FROM
	acl_simple
WHERE
	user_uuid=?
AND
	access LIKE "api:challenge:%%:write:%s"
) as challenges ON challenges.uuid = c.uuid
ORDER BY
	c.created
DESC
`
)

type SqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) ChallengeRepository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) GetChallengesByUser(userUUID string) ([]ChallengeBody, error) {
	challenges := make([]ChallengeBody, 0)
	dbItems := make([]ChallengeEntry, 0)
	// Not happy with this approach but its partly safe as we check they are logged in
	err := r.db.Select(&dbItems, fmt.Sprintf(SqlGetChallengesByUser, userUUID, userUUID), userUUID)

	if err != nil {
		return challenges, err
	}

	for _, entry := range dbItems {
		var body ChallengeBody
		json.Unmarshal([]byte(entry.Body), &body)
		body.UUID = entry.UUID
		body.Created = entry.Created.Format(time.RFC3339Nano)
		challenges = append(challenges, body)
	}

	return challenges, nil
}

func (r SqliteRepository) Join(UUID string, userUUID string) error {
	// Add user to the list
	return errors.New("TODO")
}
func (r SqliteRepository) Leave(UUID string, userUUID string) error {
	return errors.New("TODO")
}

func (r SqliteRepository) Create(challenge ChallengeEntry) error {
	_, err := r.db.Exec(SqlSaveEntry, challenge.UUID, challenge.UserUUID, challenge.Body)
	if err != nil {
		return err
	}
	return nil
}

func (r SqliteRepository) Get(UUID string) (ChallengeBody, error) {
	var body ChallengeBody
	entry := ChallengeEntry{}
	err := r.db.Get(&entry, SqlGetEntry, UUID)

	if err != nil {
		if err == sql.ErrNoRows {
			return body, ErrNotFound
		}
		return body, err
	}

	json.Unmarshal([]byte(entry.Body), &body)
	body.UUID = entry.UUID
	body.Created = entry.Created.Format(time.RFC3339Nano)
	return body, nil
}

func (r SqliteRepository) AddRecord(UUID string, extUUID string, userUUID string) error {
	_, err := r.db.Exec(SqlAddRecord, UUID, extUUID, userUUID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return nil
		}
		return err
	}
	return nil
}

func (r SqliteRepository) DeleteRecord(extUUID string, userUUID string) error {
	_, err := r.db.Exec(SqlDeleteRecord, extUUID, userUUID)
	return err
	return errors.New("TODO")
}

func (r SqliteRepository) Delete(UUID string) error {
	// Delete challenge
	// Delete records
	return errors.New("TODO")
}

package challenge

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type SqliteRepository struct {
	db *sqlx.DB
}

// Get the records
type dbRecord struct {
	UserUUID string `json:"user_uuid" db:"user_uuid"`
	Record   string `json:"body" db:"body"`
}

// users
type dbUser struct {
	UserUUID    string `db:"uuid"`
	DisplayName string `db:"display_name"`
}

var (
	SqlGetEntry    = `SELECT * FROM challenge WHERE uuid=?`
	SqlSaveEntry   = `INSERT INTO challenge(uuid, user_uuid, body) values(?, ?, ?)`
	SqlDeleteEntry = `DELETE FROM challenge WHERE uuid=?`

	SqlDeleteRecords     = `DELETE FROM challenge_records WHERE uuid=?`
	SqlAddRecord         = `INSERT INTO challenge_records(uuid, user_uuid, ext_uuid) values(?, ?, ?)`
	SqlDeleteRecord      = `DELETE FROM challenge_records WHERE ext_uuid=? AND user_uuid=?`
	SqlDeleteUserRecords = `DELETE FROM challenge_records WHERE user_uuid=?`

	SqlGetChallengeUsers = `
SELECT
	uuid,
IFNULL(json_extract(body, '$.display_name'), uuid) AS display_name
FROM
	user_info
WHERE
	uuid IN(
	SELECT
		user_uuid
	FROM
		challenge_records
	WHERE
		uuid=?
)
`
	// Tightly couple the planks with the challenges for now.
	SqlGetPlankRecords = `
SELECT
	c.user_uuid, p.body
FROM
	challenge_records AS c
INNER JOIN
	plank AS p
ON (p.uuid = c.ext_uuid AND p.user_uuid = c.user_uuid)
WHERE
	c.uuid = ?
ORDER BY
	p.created
DESC
`

	SqlGetChallengesByUser = `
SELECT
	c.uuid,
	json_extract(c.body, '$.kind') AS kind,
	json_extract(c.body, '$.description') AS description,
	c.created,
	c.user_uuid
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

func NewSqliteRepository(db *sqlx.DB) ChallengeRepository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) GetChallengesByUser(userUUID string) ([]ChallengeShortInfo, error) {
	challenges := make([]ChallengeShortInfo, 0)
	dbItems := make([]ChallengeShortInfoDB, 0)
	// Not happy with this approach but its partly safe as we check they are logged in
	err := r.db.Select(&dbItems, fmt.Sprintf(SqlGetChallengesByUser, userUUID, userUUID), userUUID)

	if err != nil {
		return challenges, err
	}

	for _, entry := range dbItems {
		info := ChallengeShortInfo{
			UUID:        entry.UUID,
			Kind:        entry.Kind,
			Description: entry.Description,
			Created:     entry.Created.Format(time.RFC3339Nano),
			CreatedBy:   entry.UserUUID,
		}
		challenges = append(challenges, info)
	}

	return challenges, nil
}

func (r SqliteRepository) Join(UUID string, userUUID string) error {
	// TODO is this function needed anymore?
	// I wonder how bad this will be VS a table with challenge_uuid, user_uuid, name
	// Remove from the list
	name := "fake"
	var path string
	findPath := `
SELECT u.path
FROM
	challenge, json_tree(challenge.body, '$.users') AS u
WHERE
	u.key='user_uuid'
AND
	u.value=?;
`
	r.db.Get(&path, findPath, userUUID)

	if path != "" {
		deleteUserByPath := `UPDATE challenge SET body=json_remove(body, ?) WHERE uuid=?;`
		_, err := r.db.Exec(deleteUserByPath, path, UUID)
		fmt.Println(err)
	}

	type dbUser struct {
		UserUUID string `json:"user_uuid"`
		Name     string `json:"name"`
	}
	// Add user to the list
	b, _ := json.Marshal(dbUser{
		UserUUID: userUUID,
		Name:     name,
	})

	userObject := string(b)
	addUser := `
UPDATE
	challenge
SET
	body=json_insert(body, "$.users[#]", json(?))
WHERE
	uuid=?
`
	_, err := r.db.Exec(addUser, userObject, UUID)
	return err
}

func (r SqliteRepository) Leave(UUID string, userUUID string) error {
	// TODO is this function needed anymore?
	// I like the code
	var path string
	findPath := `
SELECT u.path
FROM
	challenge, json_tree(challenge.body, '$.users') AS u
WHERE
	u.key='user_uuid'
AND
	u.value=?;
`
	err := r.db.Get(&path, findPath, userUUID)
	if err != nil {
		fmt.Println(err)
		return errors.New("leave.failed.finding.user")
	}

	if path != "" {
		deleteUserByPath := `UPDATE challenge SET body=json_remove(body, ?) WHERE uuid=?;`
		_, err := r.db.Exec(deleteUserByPath, path, UUID)
		fmt.Println(err)
		if err != nil {
			fmt.Println(err)
			return errors.New("leave.failed.deleting.user")
		}
	}
	return nil
}

func (r SqliteRepository) Create(userUUID string, challenge ChallengeInfo) error {
	b, _ := json.Marshal(challenge)
	_, err := r.db.Exec(SqlSaveEntry, challenge.UUID, userUUID, string(b))
	if err != nil {
		return err
	}
	return nil
}

func (r SqliteRepository) Get(UUID string) (ChallengeInfo, error) {
	var challenge ChallengeInfo
	entry := ChallengeInfoDB{}
	err := r.db.Get(&entry, SqlGetEntry, UUID)

	if err != nil {
		if err == sql.ErrNoRows {
			return challenge, ErrNotFound
		}
		return challenge, err
	}

	json.Unmarshal([]byte(entry.Body), &challenge)
	challenge.UUID = entry.UUID
	challenge.CreatedBy = entry.UserUUID
	challenge.Created = entry.Created.Format(time.RFC3339Nano)
	challenge.Records = make([]ChallengePlankRecord, 0)

	dbItems := make([]dbRecord, 0)
	r.db.Select(&dbItems, SqlGetPlankRecords, UUID)

	for _, item := range dbItems {
		var record ChallengePlankRecord
		json.Unmarshal([]byte(item.Record), &record)
		record.UserUUID = item.UserUUID

		challenge.Records = append(challenge.Records, record)
	}

	challenge.Users = make([]ChallengePlankUser, 0)
	dbChallengeUsers := make([]dbUser, 0)

	r.db.Select(&dbChallengeUsers, SqlGetChallengeUsers, UUID)

	for _, item := range dbChallengeUsers {
		challenge.Users = append(challenge.Users, ChallengePlankUser{
			UserUUID: item.UserUUID,
			Name:     item.DisplayName,
		})
	}

	return challenge, nil
}

func (r SqliteRepository) AddRecord(UUID string, extUUID string, userUUID string) (int, error) {
	_, err := r.db.Exec(SqlAddRecord, UUID, userUUID, extUUID)

	if err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") {
			return http.StatusOK, nil
		}
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func (r SqliteRepository) DeleteRecord(extUUID string, userUUID string) error {
	_, err := r.db.Exec(SqlDeleteRecord, extUUID, userUUID)
	return err
}

func (r SqliteRepository) Delete(UUID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(SqlDeleteRecords, UUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(SqlDeleteEntry, UUID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (r SqliteRepository) DeleteUser(userUUID string) error {
	_, err := r.db.Exec(SqlDeleteUserRecords, userUUID)
	return err
}

// GetUsersInfo returns users with tokens, userUUID is not unique here, as one user can have many devices
// Not sure how I feel about this
func (r SqliteRepository) GetUsersInfo(challengeUUID string, mobileApps []string) ([]ChallengeNotificationUserInfo, error) {
	// Currently hardcoded to plank.v1, first user of the challenge
	// TODO in the future, we might want to pass in "plank, remind, plank.v1, remind.v1" to get specific for the challenge
	query := `
WITH _users(user_uuid, access) AS (
SELECT
	user_uuid,
	access
FROM
	acl_simple
WHERE
	ext_uuid=?
),
_usersWithWriteAccess(user_uuid) AS (
SELECT
	user_uuid
FROM
	_users
WHERE
	access LIKE "api:challenge:%%:write:%%"
),
_usersWithDisplayName(user_uuid, display_name) AS (
	SELECT
	uuid,
	IFNULL(json_extract(body, '$.display_name'), uuid) AS display_name
FROM
	user_info
WHERE
	uuid IN(SELECT user_uuid FROM _usersWithWriteAccess)
)
SELECT
	m.user_uuid,
	m.token,
	u.display_name
FROM
	mobile_device as m
INNER JOIN
	_usersWithDisplayName AS u ON (u.user_uuid = m.user_uuid)
WHERE
	m.user_uuid IN(SELECT user_uuid FROM _usersWithDisplayName)
AND
	m.app_identifier IN (?)
`

	type dbUser struct {
		UserUUID    string `db:"user_uuid"`
		DisplayName string `db:"display_name"`
		Token       string `db:"token"`
	}

	dbItems := make([]dbUser, 0)
	users := make([]ChallengeNotificationUserInfo, 0)

	if len(mobileApps) == 0 {
		return users, nil
	}

	// This should be looked up, based on the challengeUUID
	query, args, err := sqlx.In(query, challengeUUID, mobileApps)
	fmt.Println("sqlx.In", err)
	query = r.db.Rebind(query)
	err = r.db.Select(&dbItems, query, args...)
	fmt.Println("db.Select", err)

	for _, item := range dbItems {
		users = append(users, ChallengeNotificationUserInfo{
			UserUUID:    item.UserUUID,
			DisplayName: item.DisplayName,
			Token:       item.Token,
		})
	}
	return users, nil
}

// GetUserDisplayName return empty if it doesnt exist
func (r SqliteRepository) GetUserDisplayName(uuid string) string {
	query := `
SELECT
	IFNULL(json_extract(body, '$.display_name'), "")
FROM
	user_info
WHERE
	uuid=?`
	displayName := ""
	r.db.Get(&displayName, query, uuid)
	return displayName
}

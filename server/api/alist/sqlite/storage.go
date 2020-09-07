package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/alist"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/jmoiron/sqlx"
)

// Duplicate
type AlistKV struct {
	Uuid     string `db:"uuid"`
	Body     string `db:"body"`
	UserUuid string `db:"user_uuid"`
	ListType string `db:"list_type"`
}

type GetListsByUserWithFiltersArgs struct {
	Labels   []string `db:"labels"`
	UserUuid string   `db:"user_uuid"`
	ListType string   `db:"list_type"`
}

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

func NewAlist(db *sqlx.DB) alist.DatastoreAlists {
	return &store{
		db: db,
	}
}

func (store *store) GetListsByUserWithFilters(uuid string, labels string, listType string) []alist.Alist {
	var items = []alist.Alist{}
	var row AlistKV
	filterQueryWithListTypeLookup := "list_type = :list_type"

	filterQueryWithLabelLookup := `
		uuid IN (
	SELECT
	  alist_uuid
	FROM
	  alist_labels
	WHERE
		user_uuid = :user_uuid
		AND
		label IN(:labels)
	)
`

	querySelect := `
	SELECT
	  *
	FROM
		alist_kv
	WHERE
		user_uuid = :user_uuid
	`

	filterQueryWithArgs := &GetListsByUserWithFiltersArgs{
		Labels:   strings.Split(labels, ","),
		UserUuid: uuid,
		ListType: listType,
	}
	filterQueryWith := make([]string, 0)

	if len(labels) >= 1 {
		filterQueryWith = append(filterQueryWith, filterQueryWithLabelLookup)
	}

	if listType != "" {
		filterQueryWith = append(filterQueryWith, filterQueryWithListTypeLookup)
	}

	query := querySelect
	if len(filterQueryWith) > 0 {
		query = querySelect + " AND " + strings.Join(filterQueryWith, " AND ")
	}

	query, args, err := sqlx.Named(query, filterQueryWithArgs)
	query, args, err = sqlx.In(query, args...)
	query = store.db.Rebind(query)
	rows, err := store.db.Queryx(query, args...)
	if err != nil {
		log.Println(fmt.Sprintf(i18n.InternalServerErrorTalkingToDatabase, "GetListsByUserWithFilters"))
		log.Println(err)
	}

	for rows.Next() {
		rows.StructScan(&row)
		aList := convertDbRowToAlist(row)
		items = append(items, aList)
	}

	return items
}

func (store *store) GetAlist(uuid string) (alist.Alist, error) {
	var aList alist.Alist
	row := AlistKV{}
	err := store.db.Get(&row, SqlGetItemByUUID, uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			return aList, errors.New(i18n.SuccessAlistNotFound)
		}

		log.Println(fmt.Sprintf(i18n.InternalServerErrorTalkingToDatabase, "GetAlist"))
		log.Println(err)
		return aList, err
	}

	aList = convertDbRowToAlist(row)
	return aList, nil
}

func (store *store) GetAllListsByUser(userUUID string) []alist.ShortInfo {
	lists := make([]alist.ShortInfo, 0)
	query := `
SELECT
	json_extract(body, '$.info.title') AS title,
	uuid
FROM
	alist_kv
WHERE
	user_uuid=?`

	err := store.db.Select(&lists, query, userUUID)
	if err != nil {
		fmt.Println(err)
		panic("...")
	}
	return lists
}

func (store *store) GetPublicLists() []alist.ShortInfo {
	lists := make([]alist.ShortInfo, 0)
	err := store.db.Select(&lists, SqlGetPublicLists)
	if err != nil {
		fmt.Println(err)
		panic("Failed to make public lists")
	}
	return lists
}

// TODO why aList
func (store *store) SaveAlist(method string, aList alist.Alist) (alist.Alist, error) {
	jsonBytes, err := json.Marshal(&aList)
	if err != nil {
		return aList, err
	}

	jsonAlist := string(jsonBytes)

	if method == http.MethodPost {
		// dal.Alist().Insert(aList)
		_, err = store.db.Exec(SqlInsertList, aList.Uuid, aList.Info.ListType, jsonAlist, aList.User.Uuid)
	} else {
		_, err = store.db.Exec(SqlUpdateList, aList.Info.ListType, jsonAlist, aList.User.Uuid, aList.Uuid)
	}
	return aList, nil
}

func (store *store) RemoveAlist(alistUUID string, userUUID string) error {
	_, err := store.db.Exec(SqlDeleteItemByUserAndUUID, alistUUID, userUUID)
	return err
}

// Make sure the database record for alist gets
// the correct fields attached.
// The json object saved in the db, should not be
// relied on 100% for all the fields.
func convertDbRowToAlist(row AlistKV) alist.Alist {
	var aList alist.Alist
	json.Unmarshal([]byte(row.Body), &aList)
	aList.User.Uuid = row.UserUuid
	return aList
}

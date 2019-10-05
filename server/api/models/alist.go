package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/jmoiron/sqlx"
)

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

// GetListsByUserWithFilters Filter a list and return an array of lists.
// Filter by:
// - userUUID
// - userUUID, labels
// - userUUID, listType
// - userUUID, labels, listType
// Validation
// - labels needs can be single or "," separated.
// - uuid = User.Uuid
// - listType = one of the types in alist (but if its not there, it will clearly not return anything.)
func (dal *DAL) GetListsByUserWithFilters(uuid string, labels string, listType string) []*alist.Alist {
	var items = []*alist.Alist{}
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
	query = dal.Db.Rebind(query)
	rows, err := dal.Db.Queryx(query, args...)
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

// GetAlist Get alist
func (dal *DAL) GetAlist(uuid string) (*alist.Alist, error) {
	row := AlistKV{}
	query := "SELECT * FROM alist_kv WHERE uuid = ?"
	err := dal.Db.Get(&row, query, uuid)
	if err != nil {
		log.Println(fmt.Sprintf(i18n.InternalServerErrorTalkingToDatabase, "GetAlist"))
		log.Println(err)
	}
	if err != nil {
		if err.Error() == i18n.DatabaseLookupNotFound {
			return nil, errors.New(i18n.SuccessAlistNotFound)
		}
	}

	aList := convertDbRowToAlist(row)
	return aList, nil
}

func (dal *DAL) RemoveAlist(alist_uuid string, user_uuid string) error {
	aList, err := dal.GetAlist(alist_uuid)

	if err != nil {
		return err
	}

	if aList.User.Uuid != user_uuid {
		return errors.New(i18n.InputDeleteAlistOperationOwnerOnly)
	}

	dal.RemoveLabelsForAlist(alist_uuid)
	query := `
DELETE
FROM
	alist_kv
WHERE
	uuid=?
AND
	user_uuid=?
`
	_, err = dal.Db.Exec(query, alist_uuid, user_uuid)
	if err != nil {
		log.Println(fmt.Sprintf(i18n.InternalServerErrorTalkingToDatabase, "RemoveAlist"))
		log.Println(err)
	}

	// TODO Should we trigger a cleanup of site-cache?
	dal.Acl2.DeleteList(aList.Uuid)
	return err
}

/*
If empty user, we reject.
If POST, we enforce a new uuid for the list.
If empty uuid, we reject.
If PUT, do a lookup to see if the list exists.

*/
func (dal *DAL) SaveAlist(method string, aList alist.Alist) (*alist.Alist, error) {
	var err error
	var jsonBytes []byte

	err = alist.Validate(aList)
	if err != nil {
		return nil, err
	}

	if aList.User.Uuid == "" {
		return nil, errors.New(i18n.InternalServerErrorMissingUserUuid)
	}

	// We set the uuid
	if method == http.MethodPost {
		user := &uuid.User{
			Uuid: aList.User.Uuid,
		}
		playList := uuid.NewPlaylist(user)
		aList.Uuid = playList.Uuid
	}

	if aList.Uuid == "" {
		return nil, errors.New(i18n.InternalServerErrorMissingAlistUuid)
	}
	// This really shouldnt happen, but could do if called directly.
	jsonBytes, err = json.Marshal(&aList)
	checkErr(err)
	jsonAlist := string(jsonBytes)

	if method == http.MethodPut {
		current, _ := dal.GetAlist(aList.Uuid)
		if current == nil {
			return nil, errors.New(i18n.SuccessAlistNotFound)
		}

		if current != nil {
			if current.User.Uuid != aList.User.Uuid {
				return nil, errors.New(i18n.InputSaveAlistOperationOwnerOnly)
			}
		}
		// Check if what is about to be written is the same.
		a, _ := json.Marshal(&aList)
		b, _ := json.Marshal(current)
		if string(a) == string(b) {
			return &aList, nil
		}
	}

	dal.RemoveLabelsForAlist(aList.Uuid)
	err = dal.SaveLabelsForAlist(aList)
	if err != nil {
		log.Println(err)
	}

	if method == http.MethodPost {
		//dal.Acl.CreateListRoles(aList.Uuid, aList.User.Uuid)
		dal.Acl2.MakeListPrivate(aList.Uuid, aList.User.Uuid)
		queryInsert := "INSERT INTO alist_kv(uuid, list_type, body, user_uuid) values(?, ?, ?, ?)"
		_, err = dal.Db.Exec(queryInsert, aList.Uuid, aList.Info.ListType, jsonAlist, aList.User.Uuid)
	} else {
		queryUpdate := "UPDATE alist_kv SET list_type=?, body=?, user_uuid=? WHERE uuid=?"
		_, err = dal.Db.Exec(queryUpdate, aList.Info.ListType, jsonAlist, aList.User.Uuid, aList.Uuid)
	}

	if err != nil {
		return nil, err
	}

	return &aList, nil
}

// Process the lists labels,
// We post to both the user_labels and alist_labels table.
func (dal *DAL) SaveLabelsForAlist(aList alist.Alist) error {
	// Post the labels
	var statusCode int
	var err error

	for _, label := range aList.Info.Labels {
		a := NewUserLabel(label, aList.User.Uuid)
		b := NewAlistLabel(label, aList.User.Uuid, aList.Uuid)

		statusCode, err = dal.PostUserLabel(a)
		if statusCode == http.StatusBadRequest {
			return err
		}

		statusCode, err = dal.PostAlistLabel(b)
		if statusCode == http.StatusBadRequest {
			return err
		}
	}
	return err
}

// Make sure the database record for alist gets
// the correct fields attached.
// The json object saved in the db, should not be
// relied on 100% for all the fields.
func convertDbRowToAlist(row AlistKV) *alist.Alist {
	aList := new(alist.Alist)
	json.Unmarshal([]byte(row.Body), &aList)
	aList.User.Uuid = row.UserUuid
	return aList
}

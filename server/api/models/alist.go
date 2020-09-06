package models

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
	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/jmoiron/sqlx"
)

func (dal *DAL) GetPublicLists() []alist.ShortInfo {
	query := `
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
	WHERE shared_with="public";
	`
	lists := make([]alist.ShortInfo, 0)
	err := dal.Db.Select(&lists, query)
	if err != nil {
		fmt.Println(err)
		panic("Failed to make public lists")
	}
	return lists
}

func (dal *DAL) GetAllListsByUser(userUUID string) []alist.ShortInfo {
	lists := make([]alist.ShortInfo, 0)
	query := `
SELECT
	json_extract(body, '$.info.title') AS title,
	uuid
FROM
	alist_kv
WHERE
	user_uuid=?`

	err := dal.Db.Select(&lists, query, userUUID)
	if err != nil {
		fmt.Println(err)
		panic("...")
	}
	return lists
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
func (dal *DAL) GetListsByUserWithFilters(uuid string, labels string, listType string) []alist.Alist {
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
func (dal *DAL) GetAlist(uuid string) (alist.Alist, error) {
	var aList alist.Alist
	row := AlistKV{}
	err := dal.Db.Get(&row, SQL_GET_ITEM_BY_UUID, uuid)
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

func (dal *DAL) RemoveAlist(alist_uuid string, user_uuid string) error {
	aList, err := dal.GetAlist(alist_uuid)

	if err != nil {
		return err
	}

	if aList.User.Uuid != user_uuid {
		return errors.New(i18n.InputDeleteAlistOperationOwnerOnly)
	}

	dal.Labels().RemoveLabelsForAlist(alist_uuid)
	_, err = dal.Db.Exec(SQL_DELETE_ITEM_BY_USER_AND_UUID, alist_uuid, user_uuid)
	if err != nil {
		log.Println(fmt.Sprintf(i18n.InternalServerErrorTalkingToDatabase, "RemoveAlist"))
		log.Println(err)
	}

	dal.Acl.DeleteList(aList.Uuid)
	return err
}

/*
If empty user, we reject.
If POST, we enforce a new uuid for the list.
If empty uuid, we reject.
If PUT, do a lookup to see if the list exists.
*/
func (dal *DAL) SaveAlist(method string, aList alist.Alist) (alist.Alist, error) {
	var err error
	var jsonBytes []byte
	var emptyAlist alist.Alist
	err = alist.Validate(aList)
	if err != nil {
		if err == alist.ErrorSharingNotAllowedWithFrom {
			return emptyAlist, i18n.ErrorInputSaveAlistOperationFromRestriction
		}
		return emptyAlist, err
	}

	if aList.User.Uuid == "" {
		return emptyAlist, errors.New(i18n.InternalServerErrorMissingUserUuid)
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
		return emptyAlist, errors.New(i18n.InternalServerErrorMissingAlistUuid)
	}
	// This really shouldnt happen, but could do if called directly.
	jsonBytes, err = json.Marshal(&aList)
	checkErr(err)
	jsonAlist := string(jsonBytes)

	if method == http.MethodPut {
		current, _ := dal.GetAlist(aList.Uuid)
		if current.Uuid == "" {
			return emptyAlist, errors.New(i18n.SuccessAlistNotFound)
		}

		if current.User.Uuid != aList.User.Uuid {
			return emptyAlist, errors.New(i18n.InputSaveAlistOperationOwnerOnly)
		}

		if aList.Info.From != current.Info.From {
			return emptyAlist, errors.New(i18n.InputSaveAlistOperationFromModify)
		}

		// Check if what is about to be written is the same.
		a, _ := json.Marshal(&aList)
		b, _ := json.Marshal(current)
		if string(a) == string(b) {
			return aList, nil
		}
	}

	dal.Labels().RemoveLabelsForAlist(aList.Uuid)
	err = dal.SaveLabelsForAlist(aList)
	if err != nil {
		log.Println(err)
	}

	if method == http.MethodPost {
		_, err = dal.Db.Exec(SQL_INSERT_LIST, aList.Uuid, aList.Info.ListType, jsonAlist, aList.User.Uuid)
	} else {
		_, err = dal.Db.Exec(SQL_UPDATE_LIST, aList.Info.ListType, jsonAlist, aList.User.Uuid, aList.Uuid)
	}

	if err != nil {
		return emptyAlist, err
	}

	// Set shared
	switch aList.Info.SharedWith {
	case aclKeys.SharedWithPublic:
		dal.Acl.ShareListWithPublic(aList.Uuid)
	case aclKeys.SharedWithFriends:
		dal.Acl.ShareListWithFriends(aList.Uuid)
	case aclKeys.NotShared:
		fallthrough
	default:
		dal.Acl.MakeListPrivate(aList.Uuid, aList.User.Uuid)
	}

	return aList, nil
}

// Process the lists labels,
// We post to both the user_labels and alist_labels table.
func (dal *DAL) SaveLabelsForAlist(aList alist.Alist) error {
	// Post the labels
	var statusCode int
	var err error

	for _, input := range aList.Info.Labels {
		a := label.NewUserLabel(input, aList.User.Uuid)
		b := label.NewAlistLabel(input, aList.User.Uuid, aList.Uuid)

		statusCode, err = dal.Labels().PostUserLabel(a)
		if statusCode == http.StatusBadRequest {
			return err
		}

		statusCode, err = dal.Labels().PostAlistLabel(b)
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
func convertDbRowToAlist(row AlistKV) alist.Alist {
	var aList alist.Alist
	json.Unmarshal([]byte(row.Body), &aList)
	aList.User.Uuid = row.UserUuid
	return aList
}

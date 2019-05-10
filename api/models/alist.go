package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/api/alist"
	"github.com/jmoiron/sqlx"
)

// labels needs can be single or "," separated.
func (dal *DAL) GetListsByUserAndLabels(user_uuid string, labels string) []*alist.Alist {
	var items = []*alist.Alist{}
	if labels == "" {
		return items
	}
	lookUp := strings.Split(labels, ",")

	query := `
SELECT
  body
FROM alist_kv
WHERE
  uuid IN (
SELECT
  alist_uuid
FROM
  alist_labels
WHERE
	user_uuid = ?
AND
	label IN(?)
)
`
	query, args, err := sqlx.In(query, user_uuid, lookUp)
	query = dal.Db.Rebind(query)
	rows, err := dal.Db.Query(query, args...)
	if err != nil {
		log.Println(fmt.Sprintf(InternalServerErrorTalkingToDatabase, "GetListsByUserAndLabels"))
		log.Println(err)
	}
	for rows.Next() {
		aList := new(alist.Alist)
		var body string
		rows.Scan(&body)
		json.Unmarshal([]byte(body), &aList)
		items = append(items, aList)
	}

	return items
}

// GetListsByUser Get all alists by uuid (user)
func (dal *DAL) GetListsByUser(uuid string) []*alist.Alist {
	var manyAlist []string
	query := `
SELECT
	body
FROM alist_kv
WHERE
	user_uuid = ?
`
	err := dal.Db.Select(&manyAlist, query, uuid)
	if err != nil {
		log.Println(fmt.Sprintf(InternalServerErrorTalkingToDatabase, "GetListsBy"))
		log.Println(err)
	}

	items := make([]*alist.Alist, 0)
	for _, item := range manyAlist {
		aList := new(alist.Alist)
		json.Unmarshal([]byte(item), &aList)
		items = append(items, aList)
	}
	return items
}

// GetAlist Get alist
func (dal *DAL) GetAlist(uuid string) (*alist.Alist, error) {
	var body []string
	query := "SELECT body FROM alist_kv WHERE uuid = ?"
	err := dal.Db.Select(&body, query, uuid)
	if err != nil {
		log.Println(fmt.Sprintf(InternalServerErrorTalkingToDatabase, "GetAlist"))
		log.Println(err)
	}

	if len(body) == 0 {
		return nil, errors.New(SuccessAlistNotFound)
	}

	aList := new(alist.Alist)
	json.Unmarshal([]byte(body[0]), &aList)
	return aList, nil
}

func (dal *DAL) RemoveAlist(alist_uuid string, user_uuid string) error {
	dal.RemoveLabelsForAlist(alist_uuid)
	query := `
DELETE
FROM
	alist_kv
WHERE
	uuid=$1
AND
	user_uuid=$2
`
	tx := dal.Db.MustBegin()
	tx.MustExec(query, alist_uuid, user_uuid)
	err := tx.Commit()
	if err != nil {
		log.Println(fmt.Sprintf(InternalServerErrorTalkingToDatabase, "RemoveAlist"))
		log.Println(err)
	}
	return err
}

func (dal *DAL) SaveAlist(aList alist.Alist) error {
	var err error
	var jsonBytes []byte

	err = alist.Validate(aList)
	if err != nil {
		return err
	}

	if aList.Uuid == "" {
		return errors.New(InternalServerErrorMissingAlistUuid)
	}

	if aList.User.Uuid == "" {
		return errors.New(InternalServerErrorMissingUserUuid)
	}

	jsonBytes, err = json.Marshal(&aList)
	checkErr(err)
	jsonAlist := string(jsonBytes)

	current, _ := dal.GetAlist(aList.Uuid)
	dal.RemoveLabelsForAlist(aList.Uuid)
	err = dal.SaveLabelsForAlist(aList)
	if err != nil {
		log.Println(err)
	}

	tx := dal.Db.MustBegin()
	if current == nil {
		queryInsert := "INSERT INTO alist_kv(uuid, list_type, body, user_uuid) values($1, $2, $3, $4)"
		tx.MustExec(queryInsert, aList.Uuid, aList.Info.ListType, jsonAlist, aList.User.Uuid)
	} else {
		queryUpdate := "UPDATE alist_kv SET list_type=$1, body=$2, user_uuid=$3 WHERE uuid=$4"
		tx.MustExec(queryUpdate, aList.Info.ListType, jsonAlist, aList.User.Uuid, aList.Uuid)
	}

	err = tx.Commit()
	return err
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

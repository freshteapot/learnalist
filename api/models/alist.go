package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/freshteapot/learnalist-api/api/alist"
	"github.com/jmoiron/sqlx"
)

// labels needs can be single or "," separated.
func (dal *DAL) GetListsByUserAndLabels(user_uuid string, labels string) ([]*alist.Alist, error) {
	var items = []*alist.Alist{}
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

	for rows.Next() {
		aList := new(alist.Alist)
		var body string
		err = rows.Scan(&body)
		json.Unmarshal([]byte(body), &aList)
		items = append(items, aList)
	}

	return items, err
}

// GetListsBy Get all alists by uuid
func (dal *DAL) GetListsBy(uuid string) ([]*alist.Alist, error) {
	// @todo use userid and not return all lists.
	stmt, err := dal.Db.Prepare("SELECT uuid FROM alist_kv WHERE user_uuid=?")
	rows, err := stmt.Query(uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*alist.Alist, 0)
	for rows.Next() {
		var uuid string
		err = rows.Scan(&uuid)
		if err != nil {
			return nil, err
		}
		var item *alist.Alist
		item, err = dal.GetAlist(uuid)

		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

// GetAlist Get alist
func (dal *DAL) GetAlist(uuid string) (*alist.Alist, error) {
	stmt, err := dal.Db.Prepare("SELECT body FROM alist_kv WHERE uuid = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	var body string

	aList := new(alist.Alist)
	err = stmt.QueryRow(uuid).Scan(&body)
	json.Unmarshal([]byte(body), &aList)

	if err != nil {
		return nil, err
	}

	return aList, nil
}

func (dal *DAL) RemoveAlist(uuid string) error {
	var err error
	var stmt *sql.Stmt
	// @todo lock this down to the user as well.
	stmt, err = dal.Db.Prepare("DELETE FROM alist_kv WHERE uuid=?")
	checkErr(err)

	_, err = stmt.Exec(uuid)
	checkErr(err)
	return nil
}

// TODO https://github.com/freshteapot/learnalist-api/issues/20
func (dal *DAL) SaveAlist(aList alist.Alist) error {
	var err error
	var jsonBytes []byte

	err = alist.Validate(aList)
	if err != nil {
		return err
	}

	if aList.Uuid == "" {
		return errors.New("Uuid is missing, possibly an internal error")
	}

	if aList.User.Uuid == "" {
		return errors.New("User.Uuid is missing, possibly an internal error")
	}

	jsonBytes, err = json.Marshal(&aList)
	checkErr(err)
	jsonAlist := string(jsonBytes)

	current, _ := dal.GetAlist(aList.Uuid)
	dal.RemoveLabelsForAlist(aList.Uuid)
	aList = dal.SaveLabelsForAlist(aList)

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
func (dal *DAL) SaveLabelsForAlist(aList alist.Alist) alist.Alist {
	// Post the labels
	cleanLabels := make([]string, 0)
	for _, label := range aList.Info.Labels {
		if label == "" {
			continue
		}
		if len(label) > 20 {
			continue
		}
		cleanLabels = append(cleanLabels, label)
		a := NewUserLabel(label, aList.User.Uuid)
		b := NewAlistLabel(label, aList.User.Uuid, aList.Uuid)
		dal.PostUserLabel(a)
		dal.PostAlistLabel(b)
	}
	aList.Info.Labels = cleanLabels
	return aList
}

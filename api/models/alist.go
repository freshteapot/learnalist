package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"

	"github.com/freshteapot/learnalist-api/api/alist"
)

// TODO
func (dal *DAL) GetListsByUserAndLabel(uuid string, label string) ([]*alist.Alist, error) {
	return nil, nil
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
	// TODO remove labels not on the new
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

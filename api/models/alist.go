package models

import (
	"database/sql"
	"encoding/json"
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

// PostAlist Process user data and store as new in the db.
func (dal *DAL) PostAlist(uuid string, aList alist.Alist) error {
	var err error
	var jsonBytes []byte
	var jsonAlist string
	var stmt *sql.Stmt

	jsonBytes, err = json.Marshal(&aList)
	checkErr(err)
	jsonAlist = string(jsonBytes)

	stmt, err = dal.Db.Prepare("INSERT INTO alist_kv(uuid, list_type, body, user_uuid) values(?,?,?,?)")
	checkErr(err)

	_, err = stmt.Exec(uuid, aList.Info.ListType, jsonAlist, aList.User.Uuid)
	checkErr(err)
	return nil
}

// UpdateAlist Process user data and store in db as an update.
func (dal *DAL) UpdateAlist(aList alist.Alist) error {
	var err error
	var jsonBytes []byte
	var stmt *sql.Stmt

	jsonBytes, err = json.Marshal(&aList)
	checkErr(err)
	jsonAlist := string(jsonBytes)

	stmt, err = dal.Db.Prepare("UPDATE alist_kv SET list_type=?, body=? WHERE uuid=?")
	checkErr(err)

	_, err = stmt.Exec(aList.Info.ListType, jsonAlist, aList.Uuid)
	checkErr(err)
	return nil
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

// TODO
func (dal *DAL) SaveAlist(aList alist.Alist) error {
	current, _ := dal.GetAlist(aList.Uuid)
	if current == nil {
		dal.PostAlist(aList.Uuid, aList)
	} else {
		dal.UpdateAlist(aList)
	}

	return nil
}

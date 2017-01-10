package models

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/freshteapot/learnalist/api/alist"
	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

// GetListsBy Get all alists by uuid
func (dal *DAL) GetListsBy(uuid string) ([]*alist.Alist, error) {
	// @todo use userid and not return all lists.
	rows, err := dal.Db.Query("SELECT uuid FROM alist_kv")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*alist.Alist

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

	stmt, err = dal.Db.Prepare("INSERT INTO alist_kv(uuid, list_type, body) values(?,?,?)")
	checkErr(err)

	_, err = stmt.Exec(uuid, aList.Info.ListType, jsonAlist)
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

	stmt, err = dal.Db.Prepare("DELETE FROM alist_kv WHERE uuid=?")
	checkErr(err)

	_, err = stmt.Exec(uuid)
	checkErr(err)
	return nil
}

// CreateDBStructure Create the database tables
func (dal *DAL) CreateDBStructure() {
	var query string
	var err error
	query = "create table alist_kv (uuid CHARACTER(36)  not null primary key, list_type CHARACTER(3), body text);"
	_, err = dal.Db.Exec(query)
	if err != nil {
		// table alist already exists
		return
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

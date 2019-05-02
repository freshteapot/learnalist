package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"

	"github.com/freshteapot/learnalist-api/api/alist"
	"github.com/freshteapot/learnalist-api/api/authenticate"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

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

	stmt, err = dal.Db.Prepare("DELETE FROM alist_kv WHERE uuid=?")
	checkErr(err)

	_, err = stmt.Exec(uuid)
	checkErr(err)
	return nil
}

func (dal *DAL) InsertNewUser(c echo.Context) (*uuid.User, error) {
	var hash string
	var loginUser *authenticate.LoginUser
	var err error
	var stmt *sql.Stmt

	var savedUuid string
	var savedHash string
	var savedUsername string

	loginUser = &authenticate.LoginUser{}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	json.Unmarshal(jsonBytes, loginUser)
	hash, err = authenticate.HashIt(*loginUser)

	// Make sure user is unique.
	stmt, err = dal.Db.Prepare("SELECT uuid, hash, username FROM user WHERE username = ?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_ = stmt.QueryRow(loginUser.Username).Scan(&savedUuid, &savedHash, &savedUsername)

	if savedUsername != "" {
		user := &uuid.User{}
		if savedHash != hash {
			err = errors.New("Failed to save.")
			return user, err
		}
		user.Uuid = savedUuid
		return user, nil
	}

	newUser := uuid.NewUser()
	stmt, err = dal.Db.Prepare("INSERT INTO user(uuid, hash, username) values(?,?,?)")
	checkErr(err)

	_, err = stmt.Exec(newUser.Uuid, hash, loginUser.Username)
	checkErr(err)
	return &newUser, nil
}

func (dal *DAL) GetUserByCredentials(loginUser authenticate.LoginUser) (*uuid.User, error) {
	hash, _ := authenticate.HashIt(loginUser)
	stmt, err := dal.Db.Prepare("SELECT uuid FROM user WHERE username=? AND hash=?")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	user := &uuid.User{}
	err = stmt.QueryRow(loginUser.Username, hash).Scan(&user.Uuid)
	return user, err
}

// CreateDBStructure Create the database tables
func (dal *DAL) CreateDBStructure() {
	var query string

	query = "create table alist_kv (uuid CHARACTER(36)  not null primary key, list_type CHARACTER(3), body text, user_uuid CHARACTER(36));"
	_, _ = dal.Db.Exec(query)

	query = "create table user (uuid CHARACTER(36) not null primary key, hash CHARACTER(20), username text);"
	_, _ = dal.Db.Exec(query)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

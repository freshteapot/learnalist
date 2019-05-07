package models

import (
	"database/sql"
	"errors"
	"log"

	"github.com/freshteapot/learnalist-api/api/authenticate"
	"github.com/freshteapot/learnalist-api/api/uuid"
	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

func (dal *DAL) InsertNewUser(loginUser authenticate.LoginUser) (*uuid.User, error) {
	var hash string

	var err error
	var stmt *sql.Stmt

	var savedUuid string
	var savedHash string
	var savedUsername string

	hash, err = authenticate.HashIt(loginUser)

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

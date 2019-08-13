package models

import (
	"errors"
	"fmt"
	"log"

	"github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
)

type DatabaseUser struct {
	Uuid     string `db:"uuid"`
	Username string `db:"username"`
	Hash     string `db:"hash"`
}

func NewUser(username string, hash string) *DatabaseUser {
	newUser := uuid.NewUser()
	user := &DatabaseUser{
		Uuid:     newUser.Uuid,
		Hash:     hash,
		Username: username,
	}
	return user
}

func (dal *DAL) InsertNewUser(loginUser authenticate.LoginUser) (*uuid.User, error) {
	var hash string
	var err error

	hash, err = authenticate.HashIt(loginUser)
	newUser := NewUser(loginUser.Username, hash)
	query := "INSERT INTO user(uuid, hash, username) values(:uuid,:hash,:username);"

	_, err = dal.Db.NamedExec(query, newUser)
	if err != nil {
		if err != nil {
			if err.Error() == "UNIQUE constraint failed: user.username" {
				return nil, errors.New(i18n.UserInsertUsernameExists)
			}
			// This is ugly
			checkErr(err)
		}
	}

	user := &uuid.User{
		Uuid: newUser.Uuid,
	}
	return user, nil
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

func (dal *DAL) UserExists(uuid string) bool {
	var id int
	query := "SELECT 1 FROM user WHERE uuid = ?"
	err := dal.Db.Get(&id, query, uuid)
	if err != nil {
		log.Println(fmt.Sprintf(i18n.InternalServerErrorTalkingToDatabase, "UserExists"))
		log.Println(err)
	}

	if id == 1 {
		return true
	}
	return false
}

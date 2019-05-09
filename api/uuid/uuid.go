package uuid

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

type Info interface {
	ToString() string
}

type User struct {
	Uuid string `json:"uuid"`
}

func (u User) ToString() string {
	return fmt.Sprintf("user:%s", u.Uuid)
}

type PlayList struct {
	Uuid string
	User User
}

func (p PlayList) ToString() string {
	return fmt.Sprintf("%s:playlist:%s", p.User.ToString(), p.Uuid)
}

func NewUser() User {
	u := &User{
		Uuid: GetUUID("user"),
	}
	return *u
}

func NewPlaylist(user *User) PlayList {
	p := &PlayList{
		Uuid: GetUUID("playlist"),
		User: *user,
	}
	return *p
}

func GetUUID(typeOf string) string {
	// @todo is this good enough?
	var secret = uuid.NewV4()

	// @todo Remove hardcoded learnalist
	var salt = fmt.Sprintf("learnalist.net:%s", typeOf)
	u := uuid.NewV5(secret, salt)
	return u.String()
}

package uuid

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

type Info interface {
	ToString() string
}

type User struct {
	Uuid string
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
		Uuid: getUUID(),
	}
	return *u
}

func NewPlaylist(user *User) PlayList {
	p := &PlayList{
		Uuid: getUUID(),
		User: *user,
	}
	return *p
}

func getUUID() string {
	// @todo is this good enough?
	var secret = uuid.NewV4()
	// @todo fix this
	var domain = "learnalist.net"
	u := uuid.NewV5(secret, domain)
	return u.String()
}

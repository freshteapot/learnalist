package models

import (
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/jmoiron/sqlx"
)

// DB allowing us to build an abstraction layer
type DAL struct {
	Db          *sqlx.DB
	Acl         acl.Acl
	userSession user.Session
	userFromIDP user.UserFromIDP
}

func NewDAL(db *sqlx.DB, acl acl.Acl, userSession user.Session, userFromIDP user.UserFromIDP) *DAL {
	dal := &DAL{
		Db:          db,
		Acl:         acl,
		userSession: userSession,
		userFromIDP: userFromIDP,
	}
	return dal
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

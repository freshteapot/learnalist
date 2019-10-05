package models

import (
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/jmoiron/sqlx"
)

// DB allowing us to build an abstraction layer
type DAL struct {
	Db   *sqlx.DB
	Acl  acl.Acl
	Acl2 acl.Acl
}

func NewDAL(db *sqlx.DB, acl acl.Acl) *DAL {
	dal := &DAL{
		Db:   db,
		Acl:  acl,
		Acl2: acl,
	}
	return dal
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

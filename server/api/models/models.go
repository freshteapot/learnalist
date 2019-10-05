package models

import (
	"github.com/freshteapot/learnalist-api/server/api/acl"
	acl2 "github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/jmoiron/sqlx"
)

// DB allowing us to build an abstraction layer
type DAL struct {
	Db   *sqlx.DB
	Acl  *acl.Acl
	Acl2 acl2.Acl
}

func NewDAL(db *sqlx.DB, acl *acl.Acl, acl2 acl2.Acl) *DAL {
	dal := &DAL{
		Db:   db,
		Acl:  acl,
		Acl2: acl2,
	}
	return dal
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

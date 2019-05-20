package models

import (
	"github.com/freshteapot/learnalist-api/api/acl"
	"github.com/jmoiron/sqlx"
)

// DB allowing us to build an abstraction layer
type DAL struct {
	Db  *sqlx.DB
	Acl *acl.Acl
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

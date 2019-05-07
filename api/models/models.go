package models

import (
	"github.com/jmoiron/sqlx"
)

// DB allowing us to build an abstraction layer
type DAL struct {
	Db *sqlx.DB
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

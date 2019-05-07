package models

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

// DB allowing us to build an abstraction layer
type DAL struct {
	Db *sql.DB
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

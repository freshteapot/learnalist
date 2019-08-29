package database

import (
	"fmt"
	"io/ioutil"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

const PathToTestSqliteDb = "/tmp/test.db"

func GetTables() []string {
	tables := &[]string{
		"alist_kv",
		"user",
		"user_labels",
		"alist_labels",
		"casbin_rule",
	}
	return *tables
}

func NewTestDB() *sqlx.DB {
	dataSourceName := "file:" + PathToTestSqliteDb
	db := NewDB(dataSourceName)

	pathToDbFiles := "../../db/"
	files, err := ioutil.ReadDir(pathToDbFiles)
	checkErr(err)

	for _, f := range files {
		pathToDbFile := pathToDbFiles + f.Name()
		b, err := ioutil.ReadFile(pathToDbFile)
		checkErr(err)
		query := string(b)
		db.MustExec(query)
	}

	return db
}

// NewDB load up the database
func NewDB(dataSourceName string) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", dataSourceName)

	// Very aggressive, but clearly a problem if I cant access the database.
	checkErr(err)

	err = db.Ping()
	checkErr(err)
	return db
}

func EmptyDatabase(db *sqlx.DB) {
	tables := GetTables()
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		db.MustExec(query)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

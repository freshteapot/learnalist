package models

import (
	"io/ioutil"

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

func NewTestDB() (*sqlx.DB, error) {
	dataSourceName := "file:" + PathToTestSqliteDb
	db, _ := NewDB(dataSourceName)

	pathToDbFiles := "../db/"
	files, err := ioutil.ReadDir(pathToDbFiles)
	checkErr(err)

	for _, f := range files {
		pathToDbFile := pathToDbFiles + f.Name()
		b, err := ioutil.ReadFile(pathToDbFile)
		checkErr(err)
		query := string(b)
		db.MustExec(query)
	}

	return db, nil
}

// NewDB load up the database
func NewDB(dataSourceName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	dbRef := &DAL{
		Db: db,
	}

	return dbRef.Db, nil
}

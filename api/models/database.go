package models

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

func NewTestDB() (*sql.DB, error) {
	dataSourceName := "file:test.db?mode=memory"
	return NewDB(dataSourceName)
}

// NewDB load up the database
func NewDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
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

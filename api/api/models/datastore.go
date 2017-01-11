package models

import (
	"database/sql"

	"github.com/freshteapot/learnalist-api/api/alist"
	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

// Datastore allowing us to build an abstraction layer
type Datastore interface {
	GetListsBy(uuid string) ([]*alist.Alist, error)
	GetAlist(uuid string) (*alist.Alist, error)
	PostAlist(uuid string, aList alist.Alist) error
	UpdateAlist(aList alist.Alist) error
	RemoveAlist(uuid string) error
	CreateDBStructure()
}

// DB allowing us to build an abstraction layer
type DAL struct {
	Db *sql.DB
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
	dbRef.CreateDBStructure()
	return dbRef.Db, nil
}

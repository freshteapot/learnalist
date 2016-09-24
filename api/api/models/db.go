package models

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

// Datastore allowing us to build an abstraction layer
type Datastore interface {
	GetListsBy(uuid string) ([]*Alist, error)
	GetAlist(uuid string) (*Alist, error)
	PostAlist(interface{}) (*Alist, error)
	UpdateAlist(interface{}) (*Alist, error)
	CreateDBStructure()
}

// DB allowing us to build an abstraction layer
type DB struct {
	*sql.DB
}

// NewDB load up the database
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	dbRef := &DB{db}
	dbRef.CreateDBStructure()
	return dbRef, nil
}

package models

import (
	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

//TODO
func (dal *DAL) GetLabelsByUser(Uuid string) ([]Label, error) {
	return nil, nil
}

//TODO
func (dal *DAL) SaveLabel(label Label) error {
	return nil
}

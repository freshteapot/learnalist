package models

var dal *DAL

func resetDatabase() {
	db, _ := NewTestDB()
	dal = &DAL{
		Db: db,
	}
}

package fix

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Fixup interface {
	Key() string
}

type historySqlite struct {
	db *sqlx.DB
}

type HistoryRepository interface {
	Exists(theFix Fixup) (bool, error)
	Save(theFix Fixup) error
}

func NewHistory(db *sqlx.DB) HistoryRepository {
	return historySqlite{db: db}
}

func (h historySqlite) Save(theFix Fixup) error {
	query := "INSERT INTO fixup_history(the_fix) values(?)"
	_, err := h.db.Exec(query, theFix.Key())
	return err
}

func (h historySqlite) Exists(theFix Fixup) (bool, error) {
	fmt.Println(theFix.Key())
	var id int
	query := `SELECT 1 FROM fixup_history WHERE the_fix=?`
	err := h.db.Get(&id, query, theFix.Key())
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		// Set to true, incase people are not listening for err
		return true, err
	}

	if id != 1 {
		return false, nil
	}
	return true, nil
}

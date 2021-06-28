package payment

import (
	"database/sql"
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/jmoiron/sqlx"
	"github.com/stripe/stripe-go/v71"
)

var (
	SqlGetEvent  = `SELECT body FROM payment WHERE id=?`
	SqlSaveEvent = `INSERT INTO payment (id, body) VALUES (:id, :body);`
)

type paymentSqliteRepository struct {
	db *sqlx.DB
}

func NewSqliteRepository(db *sqlx.DB) paymentSqliteRepository {
	return paymentSqliteRepository{
		db: db,
	}
}

func (r paymentSqliteRepository) Save(event stripe.Event) error {
	b, _ := json.Marshal(event)
	_, err := r.db.Exec(
		SqlSaveEvent,
		event.ID,
		string(b),
	)
	return err
}

func (r paymentSqliteRepository) Get(ID string) (stripe.Event, error) {
	var (
		body  string
		event stripe.Event
	)

	err := r.db.Get(&body, SqlGetEvent, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = utils.ErrNotFound
		}
		return event, err
	}

	_ = json.Unmarshal([]byte(body), &event)
	return event, err

}

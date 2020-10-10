package fix

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type plankV1 struct {
	db  *sqlx.DB
	key string
}

type plankV1UserAndData struct {
	AlistUUID string `db:"alist_uuid"`
	UserUUID  string `db:"user_uuid"`
	Data      string `db:"data"`
}

func NewPlankV1(db *sqlx.DB) plankV1 {
	return plankV1{
		key: "fix-plank-v1",
		db:  db,
	}
}

func (f plankV1) Key() string {
	return f.key
}

func (f plankV1) GetLists() []string {
	db := f.db
	lists := make([]string, 0)
	query := `
SELECT
	uuid
FROM
	alist_kv,
	json_each(body, '$.info.labels')
WHERE
	json_valid(body)
AND
	json_each.value LIKE 'plank';
`
	err := db.Select(&lists, query)
	if err != nil {
		fmt.Println(err)
		panic("...")
	}
	return lists
}

func (f plankV1) GetPlankRecords() []plankV1UserAndData {
	db := f.db
	lists := make([]plankV1UserAndData, 0)
	query := `
SELECT
	uuid AS alist_uuid,
	user_uuid,
	json_extract(body, '$.data') AS data
FROM
	alist_kv,
	json_each(body, '$.info.labels')
WHERE
	json_valid(body)
AND
	json_each.value LIKE 'plank';
`
	err := db.Select(&lists, query)
	if err != nil {
		fmt.Println(err)
		panic("...")
	}
	return lists
}

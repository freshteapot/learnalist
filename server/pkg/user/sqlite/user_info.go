package sqlite

import (
	"database/sql"
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/jmoiron/sqlx"
)

type userInfoSqliteRepository struct {
	db *sqlx.DB
}

const (
	SqlUserInfoGet    = `SELECT json_extract(body, '$') AS body FROM user_info WHERE uuid=?`
	SqlUserInfoDelete = `DELETE FROM user_info WHERE uuid=?`
	SqlUserInfoSave   = `INSERT INTO user_info(uuid, body) values(?, json_insert(?)) ON CONFLICT (user_info.uuid) DO UPDATE SET body=json_replace(?)`
)

func NewUserInfo(db *sqlx.DB) user.UserInfoRepository {
	return userInfoSqliteRepository{
		db: db,
	}
}

func (r userInfoSqliteRepository) Get(userUUID string) (user.UserPreference, error) {
	var pref user.UserPreference
	var body []byte
	err := r.db.Get(&body, SqlUserInfoGet, userUUID)
	if err != nil {
		if err != sql.ErrNoRows {
			return pref, err
		}
		return pref, nil
	}

	json.Unmarshal(body, &pref)
	return pref, nil
}

func (r userInfoSqliteRepository) Delete(userUUID string) error {
	_, err := r.db.Exec(SqlUserInfoDelete, userUUID)
	if err != nil {
		return err
	}
	return nil
}

func (r userInfoSqliteRepository) Save(userUUID string, pref user.UserPreference) error {
	// Might as well set this
	pref.UserUUID = userUUID

	b, _ := json.Marshal(pref)
	body := string(b)
	_, err := r.db.Exec(
		SqlUserInfoSave,
		userUUID,
		body,
		body, // On conflict
	)
	if err != nil {
		return err
	}
	return nil
}

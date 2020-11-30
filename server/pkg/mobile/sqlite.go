package mobile

import (
	"net/http"

	"github.com/jmoiron/sqlx"
)

type SqliteRepository struct {
	db *sqlx.DB
}

var (
	SqlSave               = `INSERT INTO mobile_device(user_uuid, token) values(?, ?)`
	SqlDeleteDeviceByUser = `DELETE FROM mobile_device WHERE user_uuid=?`
)

func NewSqliteRepository(db *sqlx.DB) MobileRepository {
	return SqliteRepository{
		db: db,
	}
}

func (r SqliteRepository) SaveDeviceInfo(userUUID string, token string) (int, error) {
	_, err := r.db.Exec(SqlSave, userUUID, token)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: mobile_device.user_uuid, mobile_device.token" {
			return http.StatusOK, nil
		}
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func (r SqliteRepository) DeleteByUser(userUUID string) error {
	_, err := r.db.Exec(SqlDeleteDeviceByUser, userUUID)
	if err != nil {
		return err
	}
	return nil
}

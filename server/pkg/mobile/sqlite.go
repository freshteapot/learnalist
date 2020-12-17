package mobile

import (
	"errors"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/jmoiron/sqlx"
)

type SqliteRepository struct {
	db *sqlx.DB
}

var (
	SqlSave               = `INSERT INTO mobile_device(user_uuid, app_identifier, token) values(?, ?, ?)`
	SqlDeleteDeviceByUser = `DELETE FROM mobile_device WHERE user_uuid=?`
)

func NewSqliteRepository(db *sqlx.DB) MobileRepository {
	return SqliteRepository{
		db: db,
	}
}

// TODO Next change, lets drop in the object as a json object aside from the following
func (r SqliteRepository) SaveDeviceInfo(userUUID string, input openapi.HttpMobileRegisterInput) (int, error) {
	_, err := r.db.Exec(SqlSave, userUUID, input.AppIdentifier, input.Token)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: mobile_device.user_uuid, mobile_device.app_identifier, mobile_device.token" {
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

func (r SqliteRepository) DeleteByToken(token string) error {
	return errors.New("TODO")
}

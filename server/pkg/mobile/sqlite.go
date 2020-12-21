package mobile

import (
	"database/sql"
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
	SqlDeleteDeviceByApp  = `DELETE FROM mobile_device WHERE user_uuid=? AND app_identifier=?`
	SqlGetDeviceByToken   = `SELECT user_uuid, app_identifier, token FROM mobile_device WHERE token=?`
)

type dbDeviceInfo struct {
	UserUUID      string `db:"user_uuid"`
	AppIdentifier string `db:"app_identifier"`
	Token         string `db:"token"`
}

func NewSqliteRepository(db *sqlx.DB) MobileRepository {
	return SqliteRepository{
		db: db,
	}
}

// TODO Next change, lets drop in the object as a json object aside from the following
func (r SqliteRepository) SaveDeviceInfo(deviceInfo openapi.MobileDeviceInfo) (int, error) {
	_, err := r.db.Exec(SqlSave, deviceInfo.UserUuid, deviceInfo.AppIdentifier, deviceInfo.Token)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: mobile_device.user_uuid, mobile_device.app_identifier, mobile_device.token" {
			return http.StatusOK, nil
		}

		if err.Error() == "UNIQUE constraint failed: mobile_device.token" {
			return http.StatusUnprocessableEntity, errors.New("token already in use")
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

func (r SqliteRepository) DeleteByApp(userUUID string, appIdentifier string) error {
	_, err := r.db.Exec(SqlDeleteDeviceByApp, userUUID, appIdentifier)
	if err != nil {
		return err
	}
	return nil
}

func (r SqliteRepository) GetDevicesInfoByToken(token string) ([]openapi.MobileDeviceInfo, error) {
	var (
		dbItems []dbDeviceInfo
	)
	devices := make([]openapi.MobileDeviceInfo, 0)
	err := r.db.Select(&dbItems, SqlGetDeviceByToken, token)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNotFound
		}
		return devices, err
	}

	for _, dbItem := range dbItems {
		device := openapi.MobileDeviceInfo{
			AppIdentifier: dbItem.AppIdentifier,
			Token:         dbItem.Token,
			UserUuid:      dbItem.UserUUID,
		}

		devices = append(devices, device)
	}

	return devices, err
}

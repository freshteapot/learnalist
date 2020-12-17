package remind

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/jmoiron/sqlx"
)

//uuid = {userUUID}:{appIdentifier}
var (
	// SQL_SAVE_ITEM_AUTO_UPDATED = `INSERT INTO spaced_repetition(uuid, body, user_uuid, when_next) values(?, ?, ?, ?) ON CONFLICT (spaced_repetition.user_uuid, spaced_repetition.uuid) DO UPDATE SET body=?, when_next=?`
	SqlSave = `INSERT INTO daily_reminder_settings(user_uuid, app_identifier, body, when_next) values(?, ?, ?, ?) ON CONFLICT (daily_reminder_settings.user_uuid, daily_reminder_settings.app_identifier) DO UPDATE SET body=?, when_next=?`
)

type remindDailySettingsSqliteRepository struct {
	db *sqlx.DB
}

func NewRemindDailySettingsSqliteRepository(db *sqlx.DB) RemindDailySettingsRepository {
	return remindDailySettingsSqliteRepository{
		db: db,
	}
}

func (r remindDailySettingsSqliteRepository) Save(userUUID string, settings openapi.RemindDailySettings, whenNext string) error {
	b, _ := json.Marshal(settings)
	body := string(b)
	_, err := r.db.Exec(
		SqlSave,
		userUUID, settings.AppIdentifier, body, whenNext, // New
		body, whenNext, // On conflict
	)
	if err != nil {
		return err
	}
	return nil
}

func (r remindDailySettingsSqliteRepository) DeleteByUserUUID(userUUID string) error {
	return errors.New("TODO")
}

func (r remindDailySettingsSqliteRepository) DeleteByUserAndApp(userUUID string, appIdentifier string) error {
	return errors.New("TODO")
}

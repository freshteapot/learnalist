package remind

import (
	"encoding/json"
	"errors"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/jmoiron/sqlx"
)

var (
	SqlSave        = `INSERT INTO daily_reminder_settings(user_uuid, app_identifier, body, when_next) values(?, ?, ?, ?) ON CONFLICT (daily_reminder_settings.user_uuid, daily_reminder_settings.app_identifier) DO UPDATE SET body=?, when_next=?`
	SqlGetNext     = `SELECT * FROM daily_reminder_settings WHERE when_next=? ORDER BY when_next LIMIT 1`
	SqlWhoToRemind = `
WITH _settings(user_uuid, app_identifier, settings) AS (
	SELECT
		user_uuid,
		json_extract(body, '$.app_identifier') AS app_identifier,
		body AS settings
	FROM
		daily_reminder_settings
	WHERE
		when_next <=?
	ORDER BY when_next DESC
),
_with_medium(user_uuid, settings, medium) AS (
	SELECT
		s.user_uuid,
		s.settings,
		md.token AS medium
	FROM
		_settings AS s
	INNER JOIN mobile_device as md ON (md.user_uuid = s.user_uuid)
	WHERE
		md.app_identifier=s.app_identifier
)
SELECT * FROM _with_medium
`
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

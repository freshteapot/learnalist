package remind

import (
	"encoding/json"
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/jmoiron/sqlx"
)

var (
	SqlSetActivity        = `UPDATE daily_reminder_settings SET activity=? WHERE user_uuid=? AND app_identifier=?`
	SqlDeleteByDeviceInfo = `DELETE FROM daily_reminder_settings WHERE user_uuid=? AND app_identifier=?`
	SqlDeleteByUser       = `DELETE FROM daily_reminder_settings WHERE user_uuid=?`
	SqlSave               = `INSERT INTO daily_reminder_settings(user_uuid, app_identifier, body, when_next) values(?, ?, ?, ?) ON CONFLICT (daily_reminder_settings.user_uuid, daily_reminder_settings.app_identifier) DO UPDATE SET body=?, when_next=?, activity=0`
	SqlWhoToRemind        = `
WITH _settings(user_uuid, app_identifier, settings, activity) AS (
	SELECT
		user_uuid,
		json_extract(body, '$.app_identifier') AS app_identifier,
		body AS settings,
		activity
	FROM
		daily_reminder_settings
	WHERE
		when_next <=?
),
_with_medium(user_uuid, settings, medium, activity) AS (
	SELECT
		s.user_uuid,
		s.settings,
		md.token AS medium,
		activity
	FROM
		_settings AS s
	INNER JOIN mobile_device as md ON (md.user_uuid = s.user_uuid)
	WHERE
		md.app_identifier=s.app_identifier
)

SELECT * FROM _with_medium
UNION
SELECT user_uuid, settings, "" AS medium, activity FROM _settings
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

func (r remindDailySettingsSqliteRepository) DeleteByUser(userUUID string) error {
	_, err := r.db.Exec(SqlDeleteByUser, userUUID)
	if err != nil {
		return err
	}
	return nil
}

func (r remindDailySettingsSqliteRepository) DeleteByApp(userUUID string, appIdentifier string) error {
	_, err := r.db.Exec(SqlDeleteByDeviceInfo, userUUID, appIdentifier)
	if err != nil {
		return err
	}
	return nil
}

func (r remindDailySettingsSqliteRepository) ActivityHappened(userUUID string, appIdentifier string) error {
	_, err := r.db.Exec(SqlSetActivity, 1, userUUID, appIdentifier)
	if err != nil {
		return err
	}
	return nil
}

// WhoToRemind return users with remind set.
// Medium can be empty, which means the mobile_device has not been registered yet
func (r remindDailySettingsSqliteRepository) WhoToRemind() []RemindMe {
	type dbItem struct {
		UserUUID string `db:"user_uuid"`
		Settings string `db:"settings"`
		Medium   string `db:"medium"`
		Activity bool   `db:"activity"`
	}

	dbItems := make([]dbItem, 0)
	items := make([]RemindMe, 0)

	now := time.Now().UTC()
	whenNextTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 59, 0, now.Location())
	whenNext := whenNextTime.Format(time.RFC3339Nano)

	err := r.db.Select(&dbItems, SqlWhoToRemind, whenNext)
	if err != nil {
		panic(err)
	}

	for _, item := range dbItems {
		var settings openapi.RemindDailySettings
		json.Unmarshal([]byte(item.Settings), &settings)

		items = append(items, RemindMe{
			UserUUID: item.UserUUID,
			Settings: settings,
			Medium:   item.Medium,
			Activity: item.Activity,
		})
	}
	return items
}

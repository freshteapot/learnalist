package remind

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/jmoiron/sqlx"
)

var (
	SqlSetActivity        = `UPDATE daily_reminder_settings SET activity=? WHERE user_uuid=? AND app_identifier=?`
	SqlDeleteByDeviceInfo = `DELETE FROM daily_reminder_settings WHERE user_uuid=? AND app_identifier=?`
	SqlDeleteByUser       = `DELETE FROM daily_reminder_settings WHERE user_uuid=?`
	SqlSave               = `INSERT INTO daily_reminder_settings(user_uuid, app_identifier, body, when_next) values(?, ?, ?, ?) ON CONFLICT (daily_reminder_settings.user_uuid, daily_reminder_settings.app_identifier) DO UPDATE SET body=?, when_next=?, activity=0`
	SqlGetReminders       = `
WITH
_base(user_uuid, app_identifier, settings, activity) AS (
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
		b.user_uuid,
		b.settings,
		md.token AS medium,
		activity
	FROM
		_base AS b
	INNER JOIN mobile_device AS md ON (md.user_uuid = b.user_uuid)
	WHERE
		md.app_identifier=b.app_identifier
),
_with_or_without_medium(user_uuid, settings, medium, activity) AS (
	SELECT user_uuid, settings, "" AS medium, activity FROM _base
	UNION
	SELECT user_uuid, settings, medium, activity FROM _with_medium
)
SELECT
    JSON_OBJECT(
        'user_uuid', user_uuid,
        'medium', JSON_GROUP_ARRAY(medium),
        'settings', JSON_EXTRACT(settings, '$'),
        'activity', activity
    )
FROM
    _with_or_without_medium
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

// GetReminders return users with remind set.
// Medium can be empty, which means the mobile_device has not been registered yet
func (r remindDailySettingsSqliteRepository) GetReminders(whenNext string) ([]RemindMe, error) {
	dbItems := make([][]byte, 0)
	items := make([]RemindMe, 0)

	err := r.db.Select(&dbItems, SqlGetReminders, whenNext)
	if err != nil {
		return items, err
	}

	for _, item := range dbItems {

		var r RemindMe
		json.Unmarshal(item, &r)
		// Seems to be needed as I am now returning a json object
		if r.UserUUID == "" {
			continue
		}

		items = append(items, r)
	}
	return items, nil
}

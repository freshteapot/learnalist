package remind

import (
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	SpacedRepetitionSqlUpdateSent   = `UPDATE spaced_repetition_reminder SET sent=? WHERE user_uuid=?`
	SpacedRepetitionSqlDeleteByUser = `DELETE FROM spaced_repetition_reminder WHERE user_uuid=?`
	SpacedRepetitionSqlSave         = `INSERT INTO spaced_repetition_reminder(user_uuid, when_next, last_active) values(?, ?, ?) ON CONFLICT (spaced_repetition_reminder.user_uuid) DO UPDATE SET when_next=?, last_active=?, sent=0`
	SpacedRepetitionSqlGetReminders = `
WITH
_base(user_uuid, when_next, last_active) AS (
	SELECT
		user_uuid,
		when_next,
		last_active
	FROM
		spaced_repetition_reminder
	WHERE
		sent = 0
	AND
		when_next <= ?
	AND
		last_active <= ?
),
_with_medium(user_uuid, when_next, last_active, medium) AS (
	SELECT
		b.user_uuid,
		b.when_next,
		b.last_active,
		md.token AS medium
	FROM
		_base AS b
	INNER JOIN mobile_device AS md ON (md.user_uuid = b.user_uuid)
	WHERE
		md.app_identifier = "remind_v1"
),
_with_or_without_medium(user_uuid, when_next, last_active, medium) AS (
	SELECT user_uuid, when_next, last_active, "" AS medium FROM _base
	UNION
	SELECT user_uuid, when_next, last_active, medium FROM _with_medium
)

SELECT
    JSON_OBJECT(
		'user_uuid', user_uuid,
		'when_next', when_next,
		'last_active', last_active,
		'medium', JSON_GROUP_ARRAY(medium),
		'sent', 0
    )
FROM
	_with_or_without_medium
GROUP BY
	user_uuid
`
)

type remindSpacedRepetitionSqliteRepository struct {
	db *sqlx.DB
}

func NewRemindSpacedRepetitionSqliteRepository(db *sqlx.DB) remindSpacedRepetitionSqliteRepository {
	return remindSpacedRepetitionSqliteRepository{
		db: db,
	}
}

// SetReminder takes an upsert approach and adds a record to the system in non-sent mode
func (r remindSpacedRepetitionSqliteRepository) SetReminder(userUUID string, whenNext time.Time, lastActive time.Time) error {
	sWhenNext := whenNext.Format(time.RFC3339)
	sLastActive := lastActive.Format(time.RFC3339)

	_, err := r.db.Exec(
		SpacedRepetitionSqlSave,
		userUUID, sWhenNext, sLastActive, // New
		sWhenNext, sLastActive, // On conflict
	)

	if err != nil {
		return err
	}
	return nil
}

func (r remindSpacedRepetitionSqliteRepository) DeleteByUser(userUUID string) error {
	_, err := r.db.Exec(SpacedRepetitionSqlDeleteByUser, userUUID)
	if err != nil {
		return err
	}
	return nil
}

func (r remindSpacedRepetitionSqliteRepository) UpdateSent(userUUID string, sent int) error {
	_, err := r.db.Exec(SpacedRepetitionSqlUpdateSent, sent, userUUID)
	if err != nil {
		return err
	}
	return nil
}

// GetReminders return reminders
// Medium can be empty, which means the mobile_device has not been registered yet
func (r remindSpacedRepetitionSqliteRepository) GetReminders(whenNext string, lastActive string) ([]SpacedRepetitionReminder, error) {
	dbItems := make([][]byte, 0)
	items := make([]SpacedRepetitionReminder, 0)
	err := r.db.Select(&dbItems, SpacedRepetitionSqlGetReminders, whenNext, lastActive)
	if err != nil {
		return items, err
	}

	for _, item := range dbItems {
		var r SpacedRepetitionReminder
		json.Unmarshal(item, &r)
		// TODO with the change to group by, I am not sure we need this protection anymore.
		// Seems to be needed as I am now returning a json object
		if r.UserUUID == "" {
			continue
		}

		items = append(items, r)
	}
	return items, nil
}

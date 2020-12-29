package remind

import (
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	SpacedRepetitionSqlSetPushEnabled = `UPDATE spaced_repetition_reminder SET push_enabled=? WHERE user_uuid=?`
	SpacedRepetitionSqlUpdateSent     = `UPDATE spaced_repetition_reminder SET sent=? WHERE user_uuid=?`
	SpacedRepetitionSqlDeleteByUser   = `DELETE FROM spaced_repetition_reminder WHERE user_uuid=?`
	SpacedRepetitionSqlSave           = `INSERT INTO spaced_repetition_reminder(user_uuid, when_next, last_active) values(?, ?, ?) ON CONFLICT (spaced_repetition_reminder.user_uuid) DO UPDATE SET when_next=?, last_active=?, sent=0`
	SpacedRepetitionSqlGetReminders   = `
WITH _base(user_uuid, when_next, last_active) AS (
	SELECT
		user_uuid,
		when_next,
		last_active
	FROM
		spaced_repetition_reminder
	WHERE
		sent = 0
	AND
		push_enabled = 1
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
),
_reduce(user_uuid, when_next, last_active, medium, rk) AS (
	SELECT
	*,
	ROW_NUMBER() OVER(PARTITION BY user_uuid ORDER BY medium DESC) AS rk
		FROM _with_or_without_medium
)

SELECT user_uuid, when_next, last_active, medium FROM _reduce WHERE rk = 1
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

// SetPushEnabled specifically make it clear that we want to enable / disable push
// This might need reviewing if we offer email / push etc
func (r remindSpacedRepetitionSqliteRepository) SetPushEnabled(userUUID string, enabled int32) error {
	_, err := r.db.Exec(SpacedRepetitionSqlSetPushEnabled, enabled, userUUID)
	if err != nil {
		return err
	}
	return nil
}

// GetReminders return reminders
// Medium can be empty, which means the mobile_device has not been registered yet
func (r remindSpacedRepetitionSqliteRepository) GetReminders(whenNext string, lastActive string) ([]SpacedRepetitionReminder, error) {
	type dbItem struct {
		UserUUID   string    `db:"user_uuid"`
		WhenNext   time.Time `db:"when_next"`
		LastActive time.Time `db:"last_active"`
		Medium     string    `db:"medium"` // Token or email
	}

	dbItems := make([]dbItem, 0)
	items := make([]SpacedRepetitionReminder, 0)
	// TODO How to make this ignore users who have declined events
	err := r.db.Select(&dbItems, SpacedRepetitionSqlGetReminders, whenNext, lastActive)
	if err != nil {
		return items, err
	}

	for _, item := range dbItems {
		items = append(items, SpacedRepetitionReminder{
			UserUUID:   item.UserUUID,
			WhenNext:   item.WhenNext,
			LastActive: item.LastActive,
			Medium:     item.Medium,
			Sent:       0,
		})
	}
	return items, nil
}

package remind

import (
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
)

var (
	EventApiRemindDailySettings = "api.remind.daily.settings"
	UserPreferenceKey           = "daily_reminder"
	ReminderNotSentYet          = 0
	ReminderSent                = 1
	ReminderSkipped             = 2
)

type RemindSpacedRepetitionRepository interface {
	DeleteByUser(userUUID string) error
	SetReminder(userUUID string, whenNext time.Time, lastActive time.Time) error
	UpdateSent(userUUID string, sent int) error
	GetReminders(whenNext string, lastActive string) ([]SpacedRepetitionReminder, error)
}

type RemindDailySettingsRepository interface {
	Save(userUUID string, settings openapi.RemindDailySettings, whenNext string) error
	DeleteByUser(userUUID string) error
	DeleteByApp(userUUID string, appIdentifier string) error
	ActivityHappened(userUUID string, appIdentifier string) error
	GetReminders(whenNext string) ([]RemindMe, error)
}

type RemindMe struct {
	UserUUID string                      `json:"user_uuid"`
	Settings openapi.RemindDailySettings `json:"settings"`
	Medium   []string                    `json:"medium"`
	Activity bool                        `json:"activity"`
}

type SpacedRepetitionReminder struct {
	UserUUID   string    `json:"user_uuid"`
	WhenNext   time.Time `json:"when_next"`
	LastActive time.Time `json:"last_active"`
	Sent       int       `json:"sent"` // 0, 1, 2
	Medium     []string  `json:"medium"`
}

package remind

import (
	"errors"
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/nats-io/stan.go"
)

var (
	ErrNotFound = errors.New("not.found")
)

var (
	EventApiRemindAppSettingsRemindV1 = "api.remind.app.settings.remind_v1"
	EventApiRemindDailySettings       = "api.remind.daily.settings"
	UserPreferenceKey                 = "daily_reminder"
	ReminderNotSentYet                = 0
	ReminderSent                      = 1
	ReminderSkipped                   = 2
)

type NatsSubscriber interface {
	Subscribe(topic string, sc stan.Conn) error
	Close()
}
type RemindSpacedRepetitionRepository interface {
	DeleteByUser(userUUID string) error
	SetReminder(userUUID string, whenNext time.Time, lastActive time.Time) error
	UpdateSent(userUUID string, sent int) error
	SetPushEnabled(userUUID string, enabled int32) error
	GetReminders(whenNext string, lastActive string) ([]SpacedRepetitionReminder, error)
}

type RemindDailySettingsRepository interface {
	Save(userUUID string, settings openapi.RemindDailySettings, whenNext string) error
	DeleteByUser(userUUID string) error
	DeleteByApp(userUUID string, appIdentifier string) error
	ActivityHappened(userUUID string, appIdentifier string) error
	WhoToRemind() []RemindMe
}

type RemindMe struct {
	UserUUID string
	Settings openapi.RemindDailySettings
	Medium   string // Token or email
	Activity bool
}

type SpacedRepetitionReminder struct {
	UserUUID   string    `json:"user_uuid"`
	WhenNext   time.Time `json:"when_next"`
	LastActive time.Time `json:"last_active"`
	Sent       int       `json:"sent"`   // 0, 1, 2
	Medium     string    `json:"medium"` // Token or email
}

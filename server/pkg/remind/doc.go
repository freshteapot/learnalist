package remind

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
)

var (
	ErrNotFound = errors.New("not.found")
)

var (
	EventApiRemindDailySettings = "api.remind.daily.settings"
	UserPreferenceKey           = "daily_reminder"
)

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

package remind

import (
	"errors"
	"fmt"

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
}

type UserPreference struct {
	DailyReminder struct {
		RemindV1 *openapi.RemindDailySettings `json:"remind:v1,omitempty"` // Needed first :D
		PlankV1  *openapi.RemindDailySettings `json:"plank:v1,omitempty"`
	} `json:"daily_reminder,omitempty"`
}

func RemidDailySettingsUUID(userUUID string, appIdentifier string) string {
	return fmt.Sprintf("%s:%s", userUUID, appIdentifier)
}

type RemindMe struct {
	UserUUID string
	Settings openapi.RemindDailySettings
	Medium   string // Token or email
	Activity bool
}

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
)

type UserPreference struct {
	DailyReminder struct {
		RemindV1 *openapi.RemindDailySettings `json:"remind:v1,omitempty"` // Needed first :D
		PlankV1  *openapi.RemindDailySettings `json:"plank:v1,omitempty"`
	} `json:"daily_reminder,omitempty"`
}

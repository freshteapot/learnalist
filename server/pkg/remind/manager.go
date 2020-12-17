package remind

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/plank"
	"github.com/sirupsen/logrus"
)

type manager struct {
	settingsRepo RemindDailySettingsRepository
	mobileRepo   mobile.MobileRepository
	logContext   logrus.FieldLogger
	filterKinds  []string
}

func NewManager(settingsRepo RemindDailySettingsRepository, mobileRepo mobile.MobileRepository, logContext logrus.FieldLogger) *manager {
	return &manager{
		settingsRepo: settingsRepo,
		mobileRepo:   mobileRepo,
		logContext:   logContext,
		filterKinds: []string{
			mobile.EventMobileDeviceRemoved,
			mobile.EventMobileDeviceRegistered,
			EventApiRemindDailySettings,
			event.ApiSpacedRepetition,
			plank.EventApiPlank,
		},
	}
}

func (m *manager) FilterKindsBy() []string {
	return m.filterKinds
}

// TODO now we need token
// TODO now we need display_name
// daily_reminder_medium_push
// daily_reminder_settings
func (m *manager) Write(entry event.Eventlog) {
	switch entry.Kind {
	case mobile.EventMobileDeviceRemoved:
		fmt.Println("remove token from daily_reminder_medium_push")
	case mobile.EventMobileDeviceRegistered:
		m.processMobileDeviceRegistered(entry)
	case event.ApiSpacedRepetition:
		fmt.Println("Set event added happened=1")
	case plank.EventApiPlank:
		fmt.Println("Set event added happened=1")
	case EventApiRemindDailySettings:
		m.processSettings(entry)
	default:
		return
	}
}

func (m *manager) SendNotifications(startAt time.Time) {
	// Process queue from oldeset to newest
	// Lookup
	// SELECT * FROM daily_reminders
	// Send
	// Update + set event happened=0
	fmt.Println(startAt.Format(time.RFC3339Nano))
}

func (m *manager) whenNext(from time.Time, to time.Time) time.Time {
	whenNext := to
	if to.Before(from) {
		whenNext = to.Add(time.Duration(24 * time.Hour))
	}
	return whenNext
}

func (m *manager) processSettings(entry event.Eventlog) {
	b, _ := json.Marshal(entry.Data)
	var moment event.EventKV
	json.Unmarshal(b, &moment)
	b, _ = json.Marshal(moment.Data)
	var pref UserPreference
	json.Unmarshal(b, &pref.DailyReminder)

	userUUID := moment.UUID
	conf := pref.DailyReminder.RemindV1
	// action = delete = remove
	if entry.Action == event.ActionDeleted {
		fmt.Println("Remove settings from db")

		err := m.settingsRepo.DeleteByUserAndApp(userUUID, conf.AppIdentifier)
		if err != nil {
			m.logContext.Error("failed to remove settings")
		}
		return
	}
	// action = upsert = save
	if entry.Action == event.ActionUpsert {
		conf := pref.DailyReminder.RemindV1

		loc, _ := time.LoadLocation(conf.Tz)
		// Settings + when_next (utc)
		// Take settings
		// Calculate next entry, add it
		parts := strings.Split(conf.TimeOfDay, ":")
		hour, _ := strconv.Atoi(parts[0])
		minute, _ := strconv.Atoi(parts[1])
		now := time.Now()
		local := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		// Has the time already passed, if yes add date
		whenNext := m.whenNext(now, local)

		fmt.Printf("user:%s tz:%s timeOfDay:%s\n", userUUID, conf.Tz, conf.TimeOfDay)
		fmt.Println("is before: ", local.Before(now))
		fmt.Println("utc ", now.UTC().Format(time.RFC3339Nano))
		fmt.Println("whenNext local ", whenNext.Format(time.RFC3339Nano))
		fmt.Println("whenNext utc ", whenNext.UTC().Format(time.RFC3339Nano))
		fmt.Printf("Write config to db and set when next to %s (utc: %s)\n", whenNext.Format(time.RFC3339Nano), whenNext.UTC().Format(time.RFC3339Nano))

		// I wonder if this can be cast
		err := m.settingsRepo.Save(userUUID,
			openapi.RemindDailySettings{
				TimeOfDay:     conf.TimeOfDay,
				Tz:            conf.Tz,
				Medium:        conf.Medium,
				AppIdentifier: conf.AppIdentifier,
			},
			whenNext.UTC().Format(time.RFC3339Nano),
		)
		if err != nil {
			m.logContext.Error("failed to save settings")
		}
		return
	}
	return
}

func (m *manager) processMobileDeviceRegistered(entry event.Eventlog) {
	var momentKV event.EventKV
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &momentKV)
	b, _ = json.Marshal(momentKV.Data)
	var moment mobile.DeviceInfo
	json.Unmarshal(b, &moment)

	_, err := m.mobileRepo.SaveDeviceInfo(moment.UserUUID, openapi.HttpMobileRegisterInput{
		Token:         moment.Token,
		AppIdentifier: moment.AppIdentifier,
	})

	if err != nil {
		m.logContext.Error("failed to save mobile device")
	}
}

package remind

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/plank"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type manager struct {
	db           *sqlx.DB
	settingsRepo RemindDailySettingsRepository
	mobileRepo   mobile.MobileRepository
	logContext   logrus.FieldLogger
	filterKinds  []string
}

func NewManager(db *sqlx.DB, settingsRepo RemindDailySettingsRepository, mobileRepo mobile.MobileRepository, logContext logrus.FieldLogger) *manager {
	return &manager{
		db:           db,
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

func (m *manager) StartSendNotifications() {
	// TODO do we want to pass in context cancel?
	m.logContext.Info("Sending notifications is active")
	m.SendNotifications()
	//ticker := time.NewTicker(1 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	go func() {

		for {
			select {
			case <-ticker.C:
				m.SendNotifications()
			}
		}
	}()

	//time.Sleep(1600 * time.Millisecond)
	//ticker.Stop()
	//done <- true
	//fmt.Println("Ticker stopped")
}

func (m *manager) SendNotifications() {
	// Process queue from oldeset to newest
	// Lookup
	// SELECT * FROM daily_reminders
	// Send
	// Update + set event happened=0
	fmt.Println("looking for new notifications ", time.Now().UTC())
	reminders := m.WhoToRemind()
	if len(reminders) == 0 {
		return
	}

	title := "Daily Reminder"
	var template string

	template = "It's planking time!"
	template = "What shall we learn today"
	template = "Nice work!"
	body := template

	for _, remindMe := range reminders {
		if remindMe.Settings.AppIdentifier == apps.RemindV1 &&
			utils.StringArrayContains(remindMe.Settings.Medium, "push") {
			// Make message
			message := &messaging.Message{
				Notification: &messaging.Notification{
					Title: title,
					Body:  body,
				},
				Token: remindMe.Medium,
			}

			// Send message
			event.GetBus().Publish("notifications", event.Eventlog{
				Kind: event.KindPushNotification,
				Data: message,
			})
		}

		// Update settings when next
		m.updateSettingsWithWhenNext(remindMe.UserUUID, remindMe.Settings)
	}
}

func (m *manager) whenNext(from time.Time, to time.Time) time.Time {
	whenNext := to
	if to.Before(from) {
		whenNext = to.Add(time.Duration(24 * time.Hour))
	}
	return whenNext
}

func (m *manager) updateSettingsWithWhenNext(userUUID string, conf openapi.RemindDailySettings) error {
	loc, _ := time.LoadLocation(conf.Tz)
	parts := strings.Split(conf.TimeOfDay, ":")
	hour, _ := strconv.Atoi(parts[0])
	minute, _ := strconv.Atoi(parts[1])
	now := time.Now()
	local := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	// Has the time already passed, if yes add date
	whenNext := m.whenNext(now, local)

	return m.settingsRepo.Save(
		userUUID,
		conf,
		whenNext.UTC().Format(time.RFC3339Nano),
	)
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
		err := m.updateSettingsWithWhenNext(userUUID,
			openapi.RemindDailySettings{
				TimeOfDay:     conf.TimeOfDay,
				Tz:            conf.Tz,
				Medium:        conf.Medium,
				AppIdentifier: conf.AppIdentifier,
			})

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

func (m *manager) WhoToRemind() []RemindMe {
	type dbItem struct {
		UserUUID string `db:"user_uuid"`
		Settings string `db:"settings"`
		Medium   string `db:"medium"`
	}

	dbItems := make([]dbItem, 0)
	items := make([]RemindMe, 0)

	now := time.Now().UTC()
	whenNextTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 59, 0, now.Location())
	whenNext := whenNextTime.Format(time.RFC3339Nano)
	fmt.Println(whenNext)
	err := m.db.Select(&dbItems, SqlWhoToRemind, whenNext)
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
		})
	}
	return items
}

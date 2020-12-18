package remind

import (
	"encoding/json"
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
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/sirupsen/logrus"
)

type manager struct {
	settingsRepo RemindDailySettingsRepository
	mobileRepo   mobile.MobileRepository
	logContext   logrus.FieldLogger
	filterKinds  []string
}

func NewManager(
	settingsRepo RemindDailySettingsRepository,
	mobileRepo mobile.MobileRepository,
	logContext logrus.FieldLogger) *manager {
	return &manager{
		settingsRepo: settingsRepo,
		mobileRepo:   mobileRepo,
		logContext:   logContext,
		filterKinds: []string{
			event.MobileDeviceRemoved,
			event.MobileDeviceRegistered,
			EventApiRemindDailySettings,
			event.ApiSpacedRepetition,
			event.ApiPlank,
		},
	}
}

func (m *manager) FilterKindsBy() []string {
	return m.filterKinds
}

// TODO now we need token
// TODO now we need display_name
func (m *manager) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.MobileDeviceRemoved:
		m.processMobileDeviceRemoved(entry)
	case event.MobileDeviceRegistered:
		m.processMobileDeviceRegistered(entry)
	case event.ApiSpacedRepetition:
		// Get settings
		b, _ := json.Marshal(entry.Data)
		var moment spaced_repetition.EventSpacedRepetition
		json.Unmarshal(b, &moment)
		// TODO there is the option for other messages here
		// Ie viewed VS new
		// If we want fine controlled we need a data from the client
		if moment.Kind != spaced_repetition.EventKindNew {
			return
		}
		m.settingsRepo.ActivityHappened(moment.Data.UserUUID, apps.RemindV1)
	case event.ApiPlank:
		b, _ := json.Marshal(entry.Data)
		var moment plank.EventPlank
		json.Unmarshal(b, &moment)

		if moment.Kind != plank.EventKindNew {
			return
		}
		m.settingsRepo.ActivityHappened(moment.UserUUID, apps.PlankV1)
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
	// fmt.Println("looking for new notifications ", time.Now().UTC())
	reminders := m.settingsRepo.WhoToRemind()
	if len(reminders) == 0 {
		return
	}

	title := "Daily Reminder"
	var template string

	//template = "It's planking time!"
	msgSent := 0
	for _, remindMe := range reminders {
		if remindMe.Settings.AppIdentifier == apps.RemindV1 &&
			utils.StringArrayContains(remindMe.Settings.Medium, "push") {

			template = "What shall we learn today"
			if remindMe.Activity {
				template = "Nice work!"
			}
			body := template
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
			msgSent++
		}

		// Update settings when next
		m.updateSettingsWithWhenNext(remindMe.UserUUID, remindMe.Settings)
	}
	m.logContext.WithFields(logrus.Fields{
		"msg_sent": msgSent,
	}).Info("messages sent")
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
	userUUID := entry.UUID
	b, _ := json.Marshal(entry.Data)
	var settings openapi.RemindDailySettings
	json.Unmarshal(b, &settings)

	// action = delete = remove
	if entry.Action == event.ActionDeleted {
		err := m.settingsRepo.DeleteByApp(userUUID, settings.AppIdentifier)
		if err != nil {
			m.logContext.Error("failed to remove settings")
		}
		return
	}
	// action = upsert = save
	if entry.Action == event.ActionUpsert {
		err := m.updateSettingsWithWhenNext(userUUID, settings)

		if err != nil {
			m.logContext.Error("failed to save settings")
		}
		return
	}
	return
}

func (m *manager) processMobileDeviceRegistered(entry event.Eventlog) {
	b, _ := json.Marshal(entry.Data)
	var deviceInfo openapi.MobileDeviceInfo
	json.Unmarshal(b, &deviceInfo)

	_, err := m.mobileRepo.SaveDeviceInfo(deviceInfo)

	if err != nil {
		m.logContext.Error("failed to save mobile device")
	}
}

func (m *manager) processMobileDeviceRemoved(entry event.Eventlog) {
	b, _ := json.Marshal(entry.Data)
	var deviceInfo openapi.MobileDeviceInfo
	json.Unmarshal(b, &deviceInfo)
	err := m.mobileRepo.DeleteByApp(deviceInfo.UserUuid, deviceInfo.AppIdentifier)

	if err != nil {
		m.logContext.Error("failed to remove mobile device")
	}
}

package remind

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

type dailyManager struct {
	subscription stan.Subscription
	settingsRepo RemindDailySettingsRepository
	mobileRepo   mobile.MobileRepository
	logContext   logrus.FieldLogger
	filterKinds  []string
}

func NewDaily(
	settingsRepo RemindDailySettingsRepository,
	mobileRepo mobile.MobileRepository,
	logContext logrus.FieldLogger) *dailyManager {
	return &dailyManager{
		settingsRepo: settingsRepo,
		mobileRepo:   mobileRepo,
		logContext:   logContext,
		filterKinds: []string{
			event.MobileDeviceRemoved,
			event.MobileDeviceRegistered,
			EventApiRemindDailySettings,
			event.ApiSpacedRepetition,
			event.ApiPlank,
			event.CMDUserDelete,
			event.ApiUserDelete,
		},
	}
}

func (m *dailyManager) Subscribe(topic string, sc stan.Conn) (err error) {
	d := 200 * time.Millisecond
	// Initially we shall wait

	var timer *time.Timer
	timer = time.AfterFunc(500*time.Millisecond, func() {
		m.StartSendNotifications()
		timer.Stop()
		timer = nil
	})

	handle := func(msg *stan.Msg) {
		var moment event.Eventlog
		json.Unmarshal(msg.Data, &moment)
		if !utils.StringArrayContains(m.filterKinds, moment.Kind) {

			if timer != nil {
				timer.Reset(d)
			}

			return
		}

		m.OnEvent(moment)
		if timer != nil {
			timer.Reset(d)
		}
	}

	durableName := "remind.daily"
	m.subscription, err = sc.Subscribe(
		topic,
		handle,
		stan.DurableName(durableName),
		stan.DeliverAllAvailable(),
		stan.MaxInflight(1),
	)
	if err == nil {
		m.logContext.Info("Running")
	}
	return err
}

func (m *dailyManager) Close() {
	err := m.subscription.Close()
	if err != nil {
		m.logContext.WithField("error", err).Error("closing subscription")
	}
}

// Future might want display_name
// @event.listen: event.ApiUserDelete
// @event.listen: event.CMDUserDelete
// @event.listen: event.MobileDeviceRemoved
// @event.listen: event.MobileDeviceRegistered
// @event.listen: event.ApiSpacedRepetition
// @event.listen: event.ApiPlank
func (m *dailyManager) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiUserDelete:
		fallthrough
	case event.CMDUserDelete:
		// Delete from
		userUUID := entry.UUID
		// TODO check if empty and skip
		// Or check legacy
		logContext := m.logContext.WithFields(logrus.Fields{
			"user_uuid": userUUID,
			"event":     event.UserDeleted,
		})

		err := m.settingsRepo.DeleteByUser(userUUID)
		if err != nil {
			// Future worthy of an alert
			// TODO should we use Fatal
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("settingsRepo.DeleteByUser")
		}

		err = m.mobileRepo.DeleteByUser(userUUID)
		if err != nil {
			// Future worthy of an alert
			// TODO should we use Fatal
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("mobileRepo.DeleteByUser")
		}

		if err != nil {
			// If one of the above is missing, it should stop reminders
			return
		}

		logContext.Info("user removed")
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
		var moment event.EventPlank
		json.Unmarshal(b, &moment)

		if moment.Action != event.ActionNew {
			return
		}
		m.settingsRepo.ActivityHappened(moment.UserUUID, apps.PlankV1)
	case EventApiRemindDailySettings:
		m.processSettings(entry)
	default:
		return
	}
}

func (m *dailyManager) StartSendNotifications() {
	// TODO do we want to pass in context cancel?
	m.logContext.Info("Sending notifications is active")
	m.SendNotifications()
	//ticker := time.NewTicker(1 * time.Minute)
	ticker := time.NewTicker(5 * time.Second) // Might be too aggressive
	go func() {
		for {
			select {
			case <-ticker.C:
				m.SendNotifications()
			}
		}
	}()
}

func (m *dailyManager) shouldSendNotification(r RemindMe) bool {
	tokens := r.Medium
	// If the user doesnt have any tokens one entry will still exist
	// When empty, it means the device has not been registered
	if len(tokens) == 1 {
		if tokens[0] == "" {
			return false
		}
	}

	// RemindV1 specific rules
	if r.Settings.AppIdentifier != apps.RemindV1 {
		return false
	}

	if !utils.StringArrayContains(r.Settings.Medium, "push") {
		return false
	}
	return true
}

// @event.emit: event.KindPushNotification
func (m *dailyManager) SendNotifications() {
	reminders, err := m.settingsRepo.GetReminders(DefaultNowUTC())

	if err != nil {
		m.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Trigger restart, as I am guessing issue with the database")
	}

	if len(reminders) == 0 {
		return
	}

	// Hardcoded to only work for apps.RemindV1
	title := "Daily Reminder"

	msgSent := 0
	msgSkipped := 0
	for _, remind := range reminders {
		process := m.shouldSendNotification(remind)

		if !process {
			// We dont care if this fails, as no message would be sent
			m.updateSettingsWithWhenNext(remind.UserUUID, remind.Settings)
			msgSkipped++
			continue
		}

		template := "What shall we learn today"
		if remind.Activity {
			template = "Nice work!"
		}

		body := template
		// Loop over all the tokens attached to this user
		for _, medium := range remind.Medium {
			if medium == "" {
				continue
			}

			// Make message
			message := &messaging.Message{
				Notification: &messaging.Notification{
					Title: title,
					Body:  body,
				},
				Token: medium,
			}

			// Send message
			event.GetBus().Publish(event.TopicNotifications, event.Eventlog{
				UUID: remind.UserUUID,
				Kind: event.KindPushNotification,
				Data: message,
			})
			msgSent++
		}

		err := m.updateSettingsWithWhenNext(remind.UserUUID, remind.Settings)
		if err != nil {
			m.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Trigger restart, as I am guessing issue with the database")
		}
	}

	m.logContext.WithFields(logrus.Fields{
		"msg_sent":    msgSent,
		"msg_skipped": msgSkipped,
	}).Info("messages sent")
}

func (m *dailyManager) whenNext(from time.Time, to time.Time) time.Time {
	whenNext := to
	if to.Before(from) {
		whenNext = to.Add(time.Duration(24 * time.Hour))
	}
	return whenNext
}

func (m *dailyManager) updateSettingsWithWhenNext(userUUID string, conf openapi.RemindDailySettings) error {
	loc, _ := time.LoadLocation(conf.Tz)
	parts := strings.Split(conf.TimeOfDay, ":")
	hour, _ := strconv.Atoi(parts[0])
	minute, _ := strconv.Atoi(parts[1])
	now := time.Now()
	// TODO remove this code once in production and all settings updated?
	seconds := 0
	if len(parts) > 2 {
		seconds, _ = strconv.Atoi(parts[2])
	}

	local := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, seconds, 0, loc)
	// Has the time already passed, if yes add date
	whenNext := m.whenNext(now, local)

	return m.settingsRepo.Save(
		userUUID,
		conf,
		whenNext.UTC().Format(time.RFC3339Nano),
	)
}

func (m *dailyManager) processSettings(entry event.Eventlog) {
	userUUID := entry.UUID
	b, _ := json.Marshal(entry.Data)
	var settings openapi.RemindDailySettings
	json.Unmarshal(b, &settings)

	// action = delete = remove
	if entry.Action == event.ActionDeleted {
		err := m.settingsRepo.DeleteByApp(userUUID, settings.AppIdentifier)
		if err != nil {
			m.logContext.WithFields(logrus.Fields{
				"error":     err,
				"code_path": "manager.processSettings",
				"action":    entry.Action,
			}).Error("failed to save settings")
		}
		return
	}
	// action = upsert = save
	if entry.Action == event.ActionUpsert {
		err := m.updateSettingsWithWhenNext(userUUID, settings)

		if err != nil {
			m.logContext.WithFields(logrus.Fields{
				"error":     err,
				"code_path": "manager.processSettings",
				"action":    entry.Action,
			}).Error("failed to save settings")
		}
		return
	}
	return
}

func (m *dailyManager) processMobileDeviceRegistered(entry event.Eventlog) {
	b, _ := json.Marshal(entry.Data)
	var deviceInfo openapi.MobileDeviceInfo
	json.Unmarshal(b, &deviceInfo)

	_, err := m.mobileRepo.SaveDeviceInfo(deviceInfo)

	if err != nil {
		m.logContext.WithFields(logrus.Fields{
			"error":     err,
			"code_path": "manager.processMobileDeviceRegistered",
		}).Error("failed to save mobile device")
	}
}

func (m *dailyManager) processMobileDeviceRemoved(entry event.Eventlog) {
	b, _ := json.Marshal(entry.Data)
	var deviceInfo openapi.MobileDeviceInfo
	json.Unmarshal(b, &deviceInfo)
	err := m.mobileRepo.DeleteByApp(deviceInfo.UserUuid, deviceInfo.AppIdentifier)

	if err != nil {
		m.logContext.Error("failed to remove mobile device")
	}
}

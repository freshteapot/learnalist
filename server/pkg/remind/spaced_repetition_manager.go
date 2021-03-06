package remind

import (
	"encoding/json"
	"time"

	"firebase.google.com/go/v4/messaging"
	"github.com/freshteapot/learnalist-api/server/pkg/app_settings"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

type spacedRepetitionManager struct {
	subscription         stan.Subscription
	userInfoRepo         user.UserInfoRepository
	spacedRepetitionRepo spaced_repetition.SpacedRepetitionRepository
	remindRepo           RemindSpacedRepetitionRepository
	logContext           logrus.FieldLogger
	filterKinds          []string
}

func NewSpacedRepetition(
	userInfoRepo user.UserInfoRepository,
	spacedRepetitionRepo spaced_repetition.SpacedRepetitionRepository,
	remindRepo RemindSpacedRepetitionRepository,
	logContext logrus.FieldLogger) *spacedRepetitionManager {
	return &spacedRepetitionManager{
		userInfoRepo:         userInfoRepo,
		spacedRepetitionRepo: spacedRepetitionRepo,
		remindRepo:           remindRepo,
		logContext:           logContext,
		filterKinds: []string{
			event.ApiSpacedRepetition,
			event.CMDUserDelete,
			event.ApiUserDelete,
			event.ApiAppSettingsRemindV1,
		},
	}
}

func (m *spacedRepetitionManager) Subscribe(topic string, sc stan.Conn) (err error) {
	// Initially we shall wait
	d := 200 * time.Millisecond
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

	durableName := "remind.spacedRepetition"
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

func (m *spacedRepetitionManager) Close() {
	err := m.subscription.Close()
	if err != nil {
		m.logContext.WithField("error", err).Error("closing subscription")
	}
}

/*
	- Chris has a new record
	- Send him a message
	- Dont send him another message if activity in last 10 minutes (X).
		- event.ApiSpacedRepetition / spaced_repetition.EventKindViewed

	Events:
		- event.ApiSpacedRepetition
			- EventKindNew
			- EventKindViewed
			- EventKindDeleted
		- event.MobileDeviceRemoved
		- event.MobileDeviceRegistered
		- event.CMDUserDelete
		- event.ApiUserDelete
		- event.ApiAppSettingsRemindV1

	This is like the remind table but the context is different.
	GROUP BY (userUUID whenNext)
*/
// Run every minute and check to see who has SRS next
// Check if X time has passed
// Send message
// Update message sent
// On SRS viewed remove unsent.
// @event.listen: event.ApiAppSettingsRemindV1
// @event.listen: event.ApiSpacedRepetition
// @event.listen: event.ApiUserDelete
// @event.listen: event.CMDUserDelete
func (m *spacedRepetitionManager) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiAppSettingsRemindV1:
		userUUID := entry.UUID
		b, _ := json.Marshal(entry.Data)
		var updatedSettings openapi.AppSettingsRemindV1
		json.Unmarshal(b, &updatedSettings)

		logContext := m.logContext.WithFields(logrus.Fields{
			"event":     "spacedRepetitionManager.OnEvent",
			"kind":      entry.Kind,
			"user_uuid": userUUID,
		})

		err := app_settings.SaveRemindV1(m.userInfoRepo, userUUID, updatedSettings)
		if err != nil {
			// We might want this to be fatal
			logContext.WithFields(logrus.Fields{
				"error":  err,
				"method": "app_settings.SaveRemindV1",
			}).Fatal("Failed writing to repo")
			return
		}

		if updatedSettings.SpacedRepetition.PushEnabled == 0 {
			err := m.remindRepo.DeleteByUser(userUUID)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error":  err,
					"method": "m.remindRepo.DeleteByUser",
				}).Fatal("Failed to delete user from remind repo")
			}
			return
		}

		lastActive := time.Unix(entry.Timestamp, 0).UTC()
		m.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive)
	case event.ApiSpacedRepetition:
		b, _ := json.Marshal(entry.Data)
		var moment spaced_repetition.EventSpacedRepetition
		json.Unmarshal(b, &moment)

		lastActive := time.Unix(entry.Timestamp, 0).UTC()
		srsItem := moment.Data
		userUUID := srsItem.UserUUID

		switch moment.Kind {
		case spaced_repetition.EventKindNew:
			m.spacedRepetitionRepo.SaveEntry(srsItem)
			fallthrough
		case spaced_repetition.EventKindViewed:
			m.spacedRepetitionRepo.UpdateEntry(srsItem)

			logContext := m.logContext.WithFields(logrus.Fields{
				"event":     "spacedRepetitionManager.OnEvent",
				"kind":      moment.Kind,
				"user_uuid": userUUID,
				"uuid":      srsItem.UUID,
			})

			// Check access if no
			m.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive)
		case spaced_repetition.EventKindDeleted:
			logContext := m.logContext.WithFields(logrus.Fields{
				"event":     "spacedRepetitionManager.OnEvent",
				"kind":      moment.Kind,
				"user_uuid": userUUID,
				"uuid":      srsItem.UUID,
			})

			err := m.spacedRepetitionRepo.DeleteEntry(userUUID, srsItem.UUID)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error":  err,
					"method": "m.spacedRepetitionRepo.DeleteEntry",
				}).Fatal("Failed to delete entry")
				return
			}

			m.CheckForNextEntryAndSetReminder(logContext, userUUID, lastActive)
		}
	case event.ApiUserDelete:
		fallthrough
	case event.CMDUserDelete:
		// TODO verify delete
		userUUID := entry.UUID
		// TODO check if empty and skip
		// Or check legacy
		logContext := m.logContext.WithFields(logrus.Fields{
			"user_uuid": userUUID,
			"event":     event.UserDeleted,
			"kind":      entry.Kind,
		})

		err := m.remindRepo.DeleteByUser(userUUID)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error":  err,
				"method": "m.remindRepo.DeleteByUser",
			}).Fatal("Failed to delete user from remind repo")
		}

		// This is partly duplicated in the spaced repetition service
		err = m.spacedRepetitionRepo.DeleteByUser(userUUID)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to delete user from spacedRepetitionRepo repo")
		}

		logContext.Info("user removed")
	default:
		return
	}
}

func (m *spacedRepetitionManager) CheckForNextEntryAndSetReminder(logContext logrus.FieldLogger, userUUID string, lastActive time.Time) {
	settings, err := app_settings.GetRemindV1(m.userInfoRepo, userUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"error":  err,
				"method": "app_settings.GetRemindV1",
			}).Fatal("Failed talking to repo")
		}
	}

	if settings.SpacedRepetition.PushEnabled == 0 {
		return
	}

	nextSrsItem, err := m.spacedRepetitionRepo.GetNext(userUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			logContext.WithFields(logrus.Fields{
				"error":  err,
				"method": "m.spacedRepetitionRepo.GetNext",
			}).Fatal("Unable to get next")
			return
		}

		err := m.remindRepo.DeleteByUser(userUUID)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error":  err,
				"method": "m.remindRepo.DeleteByUser",
			}).Fatal("Failed to delete user from remind repo")
		}
		return
	}

	err = m.remindRepo.SetReminder(userUUID, nextSrsItem.WhenNext, lastActive)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "m.remindRepo.SetReminder",
		}).Fatal("Failed talking to repo")
		return
	}
}

func (m *spacedRepetitionManager) StartSendNotifications() {
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

func (m *spacedRepetitionManager) shouldSendNotification(r SpacedRepetitionReminder) bool {
	tokens := r.Medium
	// If the user doesnt have any tokens one entry will still exist
	// When empty, it means the device has not been registered
	if len(tokens) == 1 {
		if tokens[0] == "" {
			return false
		}
	}

	return true
}

func (m *spacedRepetitionManager) SendNotifications() {
	reminders, err := m.remindRepo.GetReminders(DefaultWhenNextWithLastActiveOffset())
	if err != nil {
		m.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Trigger restart, as I am guessing issue with the database")
	}

	if len(reminders) == 0 {
		return
	}

	title := "Spaced Repetition"

	msgSent := 0
	msgSkipped := 0
	for _, remind := range reminders {
		process := m.shouldSendNotification(remind)

		if !process {
			// We dont care if this fails, as no message would be sent
			m.remindRepo.UpdateSent(remind.UserUUID, ReminderSkipped)
			msgSkipped++
			continue
		}

		// This assumes push
		if process {
			body := `New entry is ready, baby steps.`

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
		}

		err := m.remindRepo.UpdateSent(remind.UserUUID, ReminderSent)
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

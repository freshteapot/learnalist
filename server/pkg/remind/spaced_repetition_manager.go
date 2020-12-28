package remind

import (
	"encoding/json"
	"time"

	"firebase.google.com/go/messaging"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

type spacedRepetitionManager struct {
	subscription         stan.Subscription
	spacedRepetitionRepo spaced_repetition.SpacedRepetitionRepository
	remindRepo           RemindSpacedRepetitionRepository
	logContext           logrus.FieldLogger
	filterKinds          []string
}

func NewSpacedRepetition(
	spacedRepetitionRepo spaced_repetition.SpacedRepetitionRepository,
	remindRepo RemindSpacedRepetitionRepository,
	logContext logrus.FieldLogger) *spacedRepetitionManager {
	return &spacedRepetitionManager{
		spacedRepetitionRepo: spacedRepetitionRepo,
		remindRepo:           remindRepo,
		logContext:           logContext,
		filterKinds: []string{
			event.ApiSpacedRepetition,
			event.CMDUserDelete,
			event.ApiUserDelete,
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

	This is like the remind table but the context is different.
	GROUP BY (userUUID whenNext)
*/
// Run every minute and check to see who has SRS next
// Check if X time has passed
// Send message
// Update message sent
// On SRS viewed remove unsent.

func (m *spacedRepetitionManager) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiSpacedRepetition:
		b, _ := json.Marshal(entry.Data)
		var moment spaced_repetition.EventSpacedRepetition
		json.Unmarshal(b, &moment)

		var lastActive time.Time
		lastActive = time.Unix(entry.Timestamp, 0).UTC()
		srsItem := moment.Data
		userUUID := srsItem.UserUUID

		switch moment.Kind {
		case spaced_repetition.EventKindNew:
			m.spacedRepetitionRepo.SaveEntry(srsItem)
			fallthrough
		case spaced_repetition.EventKindViewed:
			m.spacedRepetitionRepo.UpdateEntry(srsItem)
			// Get when_next by user_uuid
			nextSrsItem, err := m.spacedRepetitionRepo.GetNext(userUUID)
			if err != nil {
				m.logContext.WithFields(logrus.Fields{
					"user_uuid": userUUID,
				}).Error("Failed to get next, which means we were not able to set reminder")
				return
			}
			_ = m.remindRepo.SetReminder(userUUID, nextSrsItem.WhenNext, lastActive)
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
					"error": err,
				}).Error("Failed to delete entry")
				return
			}

			nextSrsItem, err := m.spacedRepetitionRepo.GetNext(userUUID)
			if err != nil {
				if err != spaced_repetition.ErrNotFound {
					logContext.WithFields(logrus.Fields{
						"error": err,
					}).Error("Unable to get next")
					return
				}

				err := m.remindRepo.DeleteByUser(userUUID)
				if err != nil {
					logContext.WithFields(logrus.Fields{
						"error": err,
					}).Error("Failed to delete user from remind repo")
				}
				return
			}
			_ = m.remindRepo.SetReminder(userUUID, nextSrsItem.WhenNext, lastActive)
		}
	case event.ApiUserDelete:
		fallthrough
	case event.CMDUserDelete:
		// TODO verify delete
		userUUID := entry.UUID
		logContext := m.logContext.WithFields(logrus.Fields{
			"user_uuid": userUUID,
			"event":     event.UserDeleted,
			"kind":      entry.Kind,
		})

		// TODO How do we delete the user?
		err := m.remindRepo.DeleteByUser(userUUID)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("Failed to delete user from remind repo")
		}
	default:
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
	for _, reminder := range reminders {
		process := true
		// When empty, it means the device has not been registered
		if reminder.Medium == "" {
			process = false
		}
		// This assumes push
		if process {
			body := `New entry is ready, baby steps.`
			// Make message
			message := &messaging.Message{
				Notification: &messaging.Notification{
					Title: title,
					Body:  body,
				},
				Token: reminder.Medium,
			}

			// Send message
			event.GetBus().Publish("notifications", event.Eventlog{
				Kind: event.KindPushNotification,
				Data: message,
			})
		}

		if process {
			err := m.remindRepo.UpdateSent(reminder.UserUUID, ReminderSent)
			if err != nil {
				m.logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Trigger restart, as I am guessing issue with the database")
			}
			msgSent++
		}

		if !process {
			m.remindRepo.UpdateSent(reminder.UserUUID, ReminderSkipped)
			msgSkipped++
		}
	}

	m.logContext.WithFields(logrus.Fields{
		"msg_sent":    msgSent,
		"msg_skipped": msgSkipped,
	}).Info("messages sent")
}

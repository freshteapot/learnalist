package challenge

import (
	"encoding/json"
	"fmt"
	"net/http"

	"firebase.google.com/go/v4/messaging"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/sirupsen/logrus"
)

// @event.listen: event.ApiUserDelete
// @event.listen: event.CMDUserDelete
// @event.listen: challenge.EventChallengeDone
func (s ChallengeService) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiUserDelete:
		fallthrough
	case event.CMDUserDelete:
		s.removeUser(entry)
		return
	case EventChallengeDone:
		s.eventChallengeDone(entry)
		return
	}
	// Not the cleanest approach
	s.eventChallengePushNotification(entry)
}

// removeUser when a user is deleted
// Currently we only remove the users entries, not any entries they created.
func (s ChallengeService) removeUser(entry event.Eventlog) {
	if !event.IsUserDeleteEvent(entry) {
		return
	}

	userUUID := entry.UUID
	_ = s.repo.DeleteUser(userUUID)
	s.logContext.WithFields(logrus.Fields{
		"user_uuid": userUUID,
		"event":     event.UserDeleted,
	}).Info("user removed")
}

// @event.emit: event.KindPushNotification
func (s ChallengeService) eventChallengeDone(entry event.Eventlog) {
	if entry.Kind != EventChallengeDone {
		return
	}

	var moment EventChallengeDoneEntry
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &moment)

	challengeUUID := moment.UUID
	if moment.Kind != EventKindPlank {
		s.logContext.WithFields(logrus.Fields{
			"kind":           moment.Kind,
			"challenge_uuid": challengeUUID,
			"user_uuid":      moment.UserUUID,
		}).Info("kind not supported, yet!")
		return
	}

	b, _ = json.Marshal(moment.Data)
	var record ChallengePlankRecord
	json.Unmarshal(b, &record)

	// Add the record
	// If it is a new entry, send a event that it was new.
	status, err := s.repo.AddRecord(challengeUUID, record.UUID, moment.UserUUID)
	if status == http.StatusInternalServerError {
		s.logContext.WithFields(logrus.Fields{
			"error":  err,
			"record": entry,
		}).Error("Failed to add record")
		return
	}

	if status != http.StatusCreated {
		s.logContext.WithFields(logrus.Fields{
			"error":  "duplicate entry",
			"record": entry,
		}).Error("Failed to add record")
		return
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: EventChallengeNewRecord,
		Data: moment,
	})

	s.updateStaticSite(ChallengeInfo{UUID: challengeUUID}, true, event.ActionUpdated)
}

// @event.emit: challenge.EventChallengeDone
func (s ChallengeService) eventChallengePushNotification(entry event.Eventlog) {
	allowed := []string{
		EventChallengeNewRecord,
		EventChallengeJoined,
		EventChallengeLeft,
	}

	if !utils.StringArrayContains(allowed, entry.Kind) {
		return
	}

	var (
		challengeUUID string
		userUUID      string
		template      string
	)

	switch entry.Kind {
	case EventChallengeNewRecord:
		var moment EventChallengeDoneEntry
		b, _ := json.Marshal(entry.Data)
		json.Unmarshal(b, &moment)
		challengeUUID = moment.UUID
		userUUID = moment.UserUUID
		entryType := "record"
		if moment.Kind == EventKindPlank {
			entryType = "plank"
		}

		template = fmt.Sprintf(`%%s added a %s to %%s`, entryType)
	case EventChallengeJoined:
		var entry2 event.EventKV
		b, _ := json.Marshal(entry.Data)
		json.Unmarshal(b, &entry2)
		b, _ = json.Marshal(entry2.Data)

		var moment ChallengeJoined
		json.Unmarshal(b, &moment)
		challengeUUID = moment.UUID
		userUUID = moment.UserUUID

		template = "%s has joined %s"
	case EventChallengeLeft:
		var entry2 event.EventKV
		b, _ := json.Marshal(entry.Data)
		json.Unmarshal(b, &entry2)
		b, _ = json.Marshal(entry2.Data)

		var moment ChallengeLeft
		json.Unmarshal(b, &moment)
		challengeUUID = moment.UUID
		userUUID = moment.UserUUID
		template = "%s has left %s"
	}

	mobileApps := make([]string, 0)
	// Possibly gets expensive as it includes the users and the results
	// Plus side of including users and results, means any summary info I want in the message is available
	// We wait for it to become an issue :P
	info, _ := s.repo.Get(challengeUUID)
	switch info.Kind {
	case KindPlankGroup:
		mobileApps = PlankGroupMobileApps
	default:
		s.logContext.WithFields(logrus.Fields{
			"kind":           info.Kind,
			"challenge_uuid": challengeUUID,
		}).Error("kind not supported for push notifications, yet!")
		return
	}

	challengeName := info.Description
	userDisplayName := s.challengePushNotificationRepository.GetUserDisplayName(userUUID)
	if userDisplayName == "" {
		userDisplayName = "Someone"
	}

	title := "Challenge update"
	body := fmt.Sprintf(template, userDisplayName, challengeName)

	users, _ := s.challengePushNotificationRepository.GetUsersInfo(challengeUUID, mobileApps)
	for _, user := range users {
		// Ignore the user who created the moment
		if user.UserUUID == userUUID {
			continue
		}

		message := &messaging.Message{
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: map[string]string{
				"uuid":   challengeUUID,
				"name":   challengeName,
				"action": entry.Kind,
			},
			Token: user.Token,
		}

		event.GetBus().Publish(event.TopicNotifications, event.Eventlog{
			UUID: userUUID,
			Kind: event.KindPushNotification,
			Data: message,
		})
	}
}

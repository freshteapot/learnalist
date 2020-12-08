package challenge

import (
	"encoding/json"
	"fmt"
	"net/http"

	"firebase.google.com/go/messaging"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

func (s ChallengeService) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiUserDelete:
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
	if entry.Kind != event.ApiUserDelete {
		return
	}

	b, err := json.Marshal(entry.Data)
	if err != nil {
		return
	}

	var moment event.EventUser
	json.Unmarshal(b, &moment)
	s.repo.DeleteUser(moment.UUID)
	s.logContext.WithFields(logrus.Fields{
		"user_uuid": moment.UUID,
		"event":     event.UserDeleted,
	}).Info("user removed")
}

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
}

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
		fmt.Println(string(b))
		template = "%s has left %s"
	}

	challengeName := s.challengeNotificationRepository.GetChallengeDescription(challengeUUID)
	userDisplayName := s.challengeNotificationRepository.GetUserDisplayName(userUUID)
	if userDisplayName == "" {
		userDisplayName = "Someone"
	}

	title := "Challenge update"
	body := fmt.Sprintf(template, userDisplayName, challengeName)

	users, _ := s.challengeNotificationRepository.GetUsersInfo(challengeUUID)
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

		event.GetBus().Publish("notifications", event.Eventlog{
			Kind: event.KindPushNotification,
			Data: message,
		})
	}
}

package challenge

import (
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

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
	// TODO how do I know when its deleted?
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

	event.GetBus().Publish(event.Eventlog{
		Kind: EventChallengeNewRecord,
		Data: moment,
	})
}

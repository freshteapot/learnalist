package challenge

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

func (s ChallengeService) eventNotify(entry event.Eventlog) {
	if entry.Kind != EventChallengeNewRecord {
		return
	}

	var moment EventChallengeDoneEntry
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &moment)

	challengeUUID := moment.UUID

	b, _ = json.Marshal(moment.Data)
	var record ChallengeRecordUUID
	json.Unmarshal(b, &record)

	// TODO move / copy to the system that sends push notifications
	// TODO copy to slack
	// TODO use this to trigger a rebuild of the challenge page for static site
	// Use this event to add user to active list
	fmt.Printf("Challenge %s (%s) has a new record %s by user %s\n",
		challengeUUID,
		moment.Kind,
		record.UUID,
		moment.UserUUID)
}

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

	fmt.Println("record", string(b))
	fmt.Println("challenge", challengeUUID)
	fmt.Println("record.uuid", record.UUID)
	fmt.Println("userUUID moment", moment.UserUUID)
	// Need to know if this event was added or ignored
	// If added trigger a new event for notifications
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

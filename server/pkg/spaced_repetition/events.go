package spaced_repetition

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

// @event.listen: event.ApiUserDelete
// @event.listen: event.CMDUserDelete
// @event.listen: event.SystemSpacedRepetition
// @event.emit: event.ApiSpacedRepetition
func (s SpacedRepetitionService) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiUserDelete:
		fallthrough
	case event.CMDUserDelete:
		s.removeUser(entry)
		return
	case event.SystemSpacedRepetition:
		s.logContext.WithFields(logrus.Fields{
			"kind": entry.Kind,
		}).Info("process event")

		b, _ := json.Marshal(entry.Data)
		var moment EventSpacedRepetition
		json.Unmarshal(b, &moment)

		if moment.Kind != EventKindNew {
			return
		}

		item := moment.Data
		err := s.repo.SaveEntry(item)
		if err != nil {
			if err != ErrSpacedRepetitionEntryExists {
				s.logContext.WithFields(logrus.Fields{
					"error":  err,
					"method": "s.OnEvent",
				}).Fatal("issue with repo")
			}
			// This I believe is used to trigger a new dripfeed
			// TODO do I handle this?
			// Why do I set the action twice?
			event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
				UUID: item.UUID,
				Kind: event.SystemSpacedRepetition,
				Data: EventSpacedRepetition{
					Kind: EventKindAlreadyInSystem,
					Data: item,
				},
				Action: EventKindAlreadyInSystem,
			})
			return
		}

		// TODO
		// I am wondering why we trigger this publish after the system
		// Is this a mistake?
		// Is this a feature
		// Why do I end up with identical events (time potenitally off)
		// The entry is a new
		// This event, raises the question of event tracing
		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: event.ApiSpacedRepetition,
			Data: EventSpacedRepetition{
				Kind: EventKindNew,
				Data: item,
			},
			Action:    event.ActionCreated,
			Timestamp: entry.Timestamp,
		})
	}
}

// removeUser when a user is deleted
func (s SpacedRepetitionService) removeUser(entry event.Eventlog) {
	if !event.IsUserDeleteEvent(entry) {
		return
	}

	userUUID := entry.UUID
	_ = s.repo.DeleteByUser(userUUID)
	s.logContext.WithFields(logrus.Fields{
		"user_uuid": userUUID,
		"event":     event.UserDeleted,
	}).Info("user removed")
}

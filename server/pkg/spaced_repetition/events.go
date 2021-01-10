package spaced_repetition

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

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
				// TODO might be too aggressive
				panic(err)
			}

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

		// The entry is a new
		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: event.ApiSpacedRepetition,
			Data: EventSpacedRepetition{
				Kind: EventKindNew,
				Data: item,
			},
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

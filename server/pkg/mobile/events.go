package mobile

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

func (s MobileService) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiUserDelete:
		s.removeUser(entry)
		return
	}
}

// removeUser when a user is deleted
// Currently we only remove the users entries, not any entries they created.
func (s MobileService) removeUser(entry event.Eventlog) {
	if entry.Kind != event.ApiUserDelete {
		return
	}

	b, err := json.Marshal(entry.Data)
	if err != nil {
		return
	}

	var moment event.EventUser
	json.Unmarshal(b, &moment)
	s.repo.DeleteByUser(moment.UUID)
	s.logContext.WithFields(logrus.Fields{
		"user_uuid": moment.UUID,
		"event":     event.UserDeleted,
	}).Info("user removed")
}

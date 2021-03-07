package acl

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/sirupsen/logrus"
)

// @event.listen: event.ApiUserRegister
// @event.listen: acl.EventPublicListAccess
func (s AclService) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiUserRegister:
		s.handleApiUserRegister(entry)
	case EventPublicListAccess:
		s.handlePublicListAccess(entry)
		return
	}
}

func (s AclService) handlePublicListAccess(entry event.Eventlog) {
	var moment EventPublicListAccessData
	b, _ := json.Marshal(entry.Data)

	err := json.Unmarshal(b, &moment)
	if err != nil {
		return
	}

	message := ""
	switch moment.Action {
	case "revoke":
		err = s.repo.RevokeUserPublicListWriteAccess(moment.UserUUID)
		message = "Access revoked"
	case "grant":
		err = s.repo.GrantUserPublicListWriteAccess(moment.UserUUID)
		message = "Access granted"
	default:
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"data": entry.Data,
	})

	if err != nil {
		logContext.Fatal("Issue talking to storage")
	}

	logContext.Info(message)
}

func (s AclService) handleApiUserRegister(entry event.Eventlog) {
	var (
		moment event.EventNewUser
		pref   user.UserPreference
	)
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &moment)

	userUUID := moment.UUID
	b, _ = json.Marshal(moment.Data)
	json.Unmarshal(b, &pref)

	if pref.Acl.PublicListWrite == 1 {
		// TODO can I rely on this pref.UserUUID?

		err := s.repo.GrantUserPublicListWriteAccess(userUUID)
		if err != nil {
			s.logContext.WithFields(logrus.Fields{
				"user_uuid":    userUUID,
				"eventHandler": "handleApiUserRegister",
				"error":        err,
				"method":       "s.repo.GrantUserPublicListWriteAccess",
			}).Fatal("Issue talking to storage")
		}
	}
}

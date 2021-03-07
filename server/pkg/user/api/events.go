package api

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/sirupsen/logrus"
)

// @event.listen: event.ApiUserRegister
// @event.listen: dripfeed.EventDripfeedAdded
// @event.listen: dripfeed.EventDripfeedRemoved
// @event.listen: dripfeed.EventDripfeedFinished
// @event.listen: acl.EventPublicListAccess
func (s UserService) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	// TODO why doesnt this handle events to delete the user?
	case event.ApiUserRegister:
		s.eventUserRegister(entry)
	case dripfeed.EventDripfeedAdded:
		b, _ := json.Marshal(entry.Data)
		var moment openapi.SpacedRepetitionOvertimeInfo
		json.Unmarshal(b, &moment)
		err := AppendAndSaveSpacedRepetition(s.userInfoRepo, moment.UserUuid, moment.AlistUuid)
		if err != nil {
			s.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("repo")
		}
	case dripfeed.EventDripfeedRemoved:
		fallthrough
	case dripfeed.EventDripfeedFinished:
		b, _ := json.Marshal(entry.Data)
		var moment openapi.SpacedRepetitionOvertimeInfo
		json.Unmarshal(b, &moment)
		// TODO handle this failing
		err := RemoveAndSaveSpacedRepetition(s.userInfoRepo, moment.UserUuid, moment.AlistUuid)
		if err != nil {
			s.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("repo")
		}
	case acl.EventPublicListAccess:
		s.handlePublicListAccess(entry)
	}
}

func (s UserService) eventUserRegister(entry event.Eventlog) {
	b, _ := json.Marshal(entry.Data)
	var moment event.EventNewUser
	json.Unmarshal(b, &moment)

	var pref user.UserPreference
	b, _ = json.Marshal(moment.Data)
	json.Unmarshal(b, &pref)

	userUUID := moment.UUID
	err := s.userInfoRepo.Save(userUUID, pref)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("repo")
	}

	if pref.Acl.PublicListWrite == 1 {
		// TODO can I rely on this pref.UserUUID?

		err := s.aclRepo.GrantUserPublicListWriteAccess(userUUID)
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

func (s UserService) handlePublicListAccess(entry event.Eventlog) {
	var moment acl.EventPublicListAccessData
	b, _ := json.Marshal(entry.Data)

	err := json.Unmarshal(b, &moment)
	if err != nil {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"event": "acl.publiclistaccess",
	})

	pref, err := s.userInfoRepo.Get(moment.UserUUID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Talking to storage")
	}

	switch moment.Action {
	case "revoke":
		pref.Acl.PublicListWrite = 0
	case "grant":
		pref.Acl.PublicListWrite = 1
	default:
		return
	}

	err = s.userInfoRepo.Save(moment.UserUUID, pref)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("Talking to storage")
	}
	logContext.Info("updated")
}

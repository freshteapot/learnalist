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

func (s MobileService) removeDeviceByToken(entry event.Eventlog) {
	// TODO this event does not exist yet
	if entry.Kind != event.MobileDeviceRemove {
		return
	}

	logEvent := "removeDevice"
	logContext := s.logContext.WithField("sub-context", logEvent)

	var momentKV event.EventKV
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &momentKV)
	token := momentKV.Data.(string)

	deviceInfo, err := s.repo.GetDeviceInfoByToken(token)
	if err != nil {
		if err == ErrNotFound {
			return
		}
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("GetDeviceInfoByToken")
		return
	}

	err = s.repo.DeleteByApp(deviceInfo.UserUuid, deviceInfo.AppIdentifier)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
			"token": token,
		}).Error("DeleteByApp")
		return
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind:   event.MobileDeviceRemoved,
		Data:   deviceInfo,
		Action: event.ActionDeleted,
	})
}

func (s MobileService) removeUser(entry event.Eventlog) {
	if entry.Kind != event.ApiUserDelete {
		return
	}
	userUUID := entry.UUID
	s.repo.DeleteByUser(userUUID)
	s.logContext.WithFields(logrus.Fields{
		"user_uuid": userUUID,
		"event":     event.UserDeleted,
	}).Info("user removed")
}

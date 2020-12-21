package mobile

import (
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

func (s MobileService) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiUserDelete:
		fallthrough
	case event.CMDUserDelete:
		s.removeUser(entry)
	case event.MobileDeviceRemove:
		s.removeDevicesByToken(entry)
	}
}

func (s MobileService) removeDevicesByToken(entry event.Eventlog) {
	if entry.Kind != event.MobileDeviceRemove {
		return
	}

	logEvent := "removeDevice"
	logContext := s.logContext.WithField("sub-context", logEvent)

	token := entry.Data.(string)
	devices, err := s.repo.GetDevicesInfoByToken(token)
	if err != nil {
		if err == ErrNotFound {
			return
		}
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("GetDevicesInfoByToken")
		return
	}

	for _, deviceInfo := range devices {
		err = s.repo.DeleteByApp(deviceInfo.UserUuid, deviceInfo.AppIdentifier)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"error": err,
				"token": token,
			}).Error("DeleteByApp")
			continue
		}

		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind:   event.MobileDeviceRemoved,
			Data:   deviceInfo,
			Action: event.ActionDeleted,
		})
	}

}

func (s MobileService) removeUser(entry event.Eventlog) {
	if !event.IsUserDeleteEvent(entry) {
		return
	}
	userUUID := entry.UUID
	s.repo.DeleteByUser(userUUID)
	s.logContext.WithFields(logrus.Fields{
		"user_uuid": userUUID,
		"event":     event.UserDeleted,
	}).Info("user removed")
}

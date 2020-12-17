package mobile

import (
	"encoding/json"
	"fmt"

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

// TODO remove device  based on token
func (s MobileService) removeDeviceByToken(entry event.Eventlog) {
	if entry.Kind != EventMobileDeviceRemove {
		return
	}
	var momentKV event.EventKV
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &momentKV)
	token := momentKV.Data.(string)
	// Lookup device?
	fmt.Printf("Remove token %s from device_table\n", token)

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: EventMobileDeviceRemoved,
		Data: event.EventKV{
			UUID: momentKV.UUID,
			Data: token,
		},
		Action: event.ActionDeleted,
	})
}

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

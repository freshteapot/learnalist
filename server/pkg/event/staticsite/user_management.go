package staticsite

import (
	"github.com/freshteapot/learnalist-api/server/pkg/event"
)

type siteManagementViaEvents struct{}

func NewSiteManagementViaEvents() siteManagementViaEvents {
	return siteManagementViaEvents{}
}

func (m siteManagementViaEvents) DeleteList(listUUID string) error {
	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.SystemListDelete,
		UUID: listUUID,
	})
	return nil
}

func (m siteManagementViaEvents) DeleteUser(userUUID string) error {
	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.SystemUserDelete,
		UUID: userUUID,
	})
	return nil
}

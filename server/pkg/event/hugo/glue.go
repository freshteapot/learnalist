package hugo

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

type glue struct {
	alistRepo  alist.DatastoreAlists
	logContext logrus.FieldLogger
}

func NewGlue(alistRepo alist.DatastoreAlists, logContext logrus.FieldLogger) glue {
	return glue{
		alistRepo:  alistRepo,
		logContext: logContext,
	}
}

func (g glue) ListenForEvents() {
	event.GetBus().Subscribe(event.TopicMonolog, "hugoGlue", g.OnEvent)
}

func (g glue) OnEvent(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiListSaved:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventList
		json.Unmarshal(b, &moment)

		userUUID := moment.UserUUID
		alistUUID := moment.UUID

		// WriteList
		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			Kind:   event.ChangesetAlistList,
			UUID:   alistUUID,
			Data:   moment.Data,
			Action: entry.Action,
		})

		// WriteListsByUser
		lists := g.alistRepo.GetAllListsByUser(userUUID)
		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			Kind:   event.ChangesetAlistUser,
			UUID:   userUUID,
			Data:   lists,
			Action: event.ActionUpdated,
		})

		// WritePublicLists
		g.triggerPublicList()
	case event.ApiUserRegister:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventNewUser
		json.Unmarshal(b, &moment)
		userUUID := moment.UUID
		lists := make([]alist.ShortInfo, 0)
		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			Kind:   event.ChangesetAlistUser,
			UUID:   userUUID,
			Data:   lists,
			Action: event.ActionCreated,
		})

	case event.SystemUserDelete:
		fallthrough
	case event.ApiUserDelete:
		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			Kind:   event.ChangesetAlistUser,
			UUID:   entry.UUID,
			Action: event.ActionDeleted,
		})

		// WritePublicLists
		g.triggerPublicList()
	case event.ApiListDelete:
		b, _ := json.Marshal(entry.Data)
		var moment event.EventList
		json.Unmarshal(b, &moment)

		userUUID := moment.UserUUID
		alistUUID := moment.UUID

		// WriteList
		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			Kind:   event.ChangesetAlistList,
			UUID:   alistUUID,
			Action: event.ActionDeleted,
		})

		// WriteListsByUser
		lists := g.alistRepo.GetAllListsByUser(userUUID)
		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			Kind:   event.ChangesetAlistUser,
			UUID:   userUUID,
			Data:   lists,
			Action: event.ActionUpdated,
		})

		// WritePublicLists
		g.triggerPublicList()
	case event.SystemListDelete:
		alistUUID := entry.UUID
		// WriteList
		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			Kind:   event.ChangesetAlistList,
			UUID:   alistUUID,
			Action: event.ActionDeleted,
		})

		// WritePublicLists
		g.triggerPublicList()
	}
}

func (g glue) triggerPublicList() {
	publicLists := g.alistRepo.GetPublicLists()
	event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
		Kind:   event.ChangesetAlistPublic,
		Data:   publicLists,
		Action: event.ActionUpdated,
	})
}

package hugo

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

func (h *HugoHelper) ListenForEvents() {
	event.GetBus().Subscribe(event.TopicStaticSite, "hugoHelper", h.OnEvent)
}

func (h *HugoHelper) OnEvent(entry event.Eventlog) {
	logContext := h.logContext.WithFields(logrus.Fields{
		"kind":   entry.Kind,
		"action": entry.Action,
	})

	switch entry.Kind {
	case event.ChangesetAlistList:
		alistUUID := entry.UUID
		logContext = logContext.WithFields(logrus.Fields{
			"alist_uuid": alistUUID,
		})

		if entry.Action == event.ActionDeleted {
			h.DeleteList(alistUUID)
			logContext.Info("to write")
			return
		}

		var aList alist.Alist
		b, _ := json.Marshal(entry.Data)
		_ = json.Unmarshal(b, &aList)

		h.WriteList(aList)
		h.ProcessContent()

		logContext.Info("to write")
	case event.ChangesetAlistUser:
		userUUID := entry.UUID
		logContext = logContext.WithFields(logrus.Fields{
			"user_uuid": userUUID,
		})
		if entry.Action == event.ActionDeleted {
			h.DeleteUser(userUUID)
			logContext.Info("to write")
			return
		}

		var lists []alist.ShortInfo

		b, _ := json.Marshal(entry.Data)
		_ = json.Unmarshal(b, &lists)
		h.WriteListsByUser(userUUID, lists)
		h.ProcessContent()
		logContext.WithFields(logrus.Fields{
			"total_lists": len(lists),
		}).Info("to write")
	case event.ChangesetAlistPublic:
		// TODO might break when public lists are too many
		var lists []alist.ShortInfo
		b, _ := json.Marshal(entry.Data)
		_ = json.Unmarshal(b, &lists)
		h.WritePublicLists(lists)
		h.ProcessContent()
		logContext.WithFields(logrus.Fields{
			"total_lists": len(lists),
		}).Info("to write")
	case event.ChangesetChallenge:
		b, _ := json.Marshal(entry.Data)
		var moment challenge.ChallengeInfo
		_ = json.Unmarshal(b, &moment)

		challengeUUID := moment.UUID
		logContext = logContext.WithFields(logrus.Fields{
			"challenge_uuid": challengeUUID,
		})

		if entry.Action == event.ActionDeleted {
			h.challengeWriter.Remove(challengeUUID)
			logContext.Info("to write")
			return
		}

		// Write challenge
		h.challengeWriter.Data(moment)
		h.challengeWriter.Content(moment)
		h.ProcessContent()
		logContext.Info("to write")
		return
	default:
		logContext.Info("not supported")
		return
	}
}
